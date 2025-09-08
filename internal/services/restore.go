/*
Copyright © 2025 Chan Alston git@chanalston.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"golang.org/x/term"
)

const JSON_METADATA_VERSION = "1.0"

type Metadata struct {
	Version     string   `json:"version"`
	Service     string   `json:"service"`
	Timestamp   string   `json:"timestamp"`
	ComposeFile string   `json:"compose_file"`
	Volumes     []Volume `json:"volumes"`
}
type Volume struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func GpgDecryptFile(encryptedContent []byte, passphrase []byte) (*crypto.VerifiedDataResult, error) {
	decHandle, err := crypto.PGP().Decryption().Password(passphrase).New()
	if err != nil {
		return nil, fmt.Errorf("unable to create decryption handle: %v", err)
	}

	decryptedContent, err := decHandle.Decrypt(encryptedContent, crypto.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %v", err)
	}

	return decryptedContent, nil
}

func PromptPassphrase() ([]byte, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.GetState(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to get terminal state: %w", err)
	}

	// Set up signal handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGHUP)
	defer signal.Stop(sigCh)

	// Restore terminal if signal received
	go func() {
		for range sigCh {
			_ = term.Restore(fd, oldState)
			os.Exit(1)
		}
	}()

	// force restore on panic
	defer func() {
		if r := recover(); r != nil {
			_ = term.Restore(fd, oldState)
			panic(r)
		}
	}()

	fmt.Fprint(os.Stderr, "Enter GPG passphrase: ")

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // move to new line after input
	if err != nil {
		return nil, fmt.Errorf("unable to read the passphrase from stdin: %v", err)
	}

	return bytePassword, nil
}

func RestoreAllDockerVolume(docker *client.Client, ctx context.Context, content []byte) error {
	data, err := ReadFileFromTarGz(content, "/backup/backup.json")
	if err != nil {
		return fmt.Errorf("unable to read the /backup/backup.json: %v", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("unable to parse the /backup/backup.json: %v", err)
	}

	for _, volumeData := range metadata.Volumes {
		err = RestoreContentToDockerVolume(docker, ctx, volumeData, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func RestoreContentToDockerVolume(docker *client.Client, ctx context.Context,
	targetVolume Volume, content []byte) error {
	// 1. Create volume
	if _, err := docker.VolumeCreate(ctx, volume.CreateOptions{Name: targetVolume.Name}); err != nil {
		return fmt.Errorf("unable to create docker volume %s: %v", targetVolume.Name, err)
	}
	fmt.Printf("Docker volume created: %s\n", targetVolume.Name)

	// 2. Create container with volume mounted

	// It is important that we trim any leading / or tar cannot find the
	// archive directory
	var command = []string{"sh", "-c", "tar -xzf - --strip-components=2 -C /restore " +
		strings.TrimLeft(targetVolume.Path, "/")}

	tempContainer, err := docker.ContainerCreate(ctx,
		&container.Config{
			Image:        "busybox:stable-glibc",
			Cmd:          command,
			Tty:          false,
			OpenStdin:    true,
			StdinOnce:    true, // <- Important: close stdin after first attach
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: targetVolume.Name,
					Target: "/restore",
				},
			},
			AutoRemove: true, // --rm
		},
		nil,
		nil,
		"", // Auto generate the name
	)
	if err != nil {
		return fmt.Errorf("unable to create a temp docker container: %v", err)
	}

	// 3. Attach to the container
	hijack, err := docker.ContainerAttach(ctx, tempContainer.ID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("unable to attach to the temp docker container: %v", err)
	}
	defer hijack.Close()

	// 4. Start the container
	if err := docker.ContainerStart(ctx, tempContainer.ID, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("unable to start to the temp docker container: %v", err)
	}

	// 5. Copy tar.gz into stdin and close properly
	go func() {
		defer func() {
			hijack.CloseWrite()
		}()

		log.Println("Copy started")
		written, err := io.Copy(hijack.Conn, bytes.NewReader(content))
		if err != nil {
			log.Printf("Error copying data: %v", err)
			return
		}
		log.Printf("Copy ended, wrote %d bytes", written)
	}()

	// 6. Read container output (important for proper cleanup)
	go func() {
		scanner := bufio.NewScanner(hijack.Reader)
		for scanner.Scan() {
			log.Println("Container output:", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading container output: %v", err)
		}
	}()

	// 7. Wait for container with timeout
	statusCh, errCh := docker.ContainerWait(ctx, tempContainer.ID, container.WaitConditionNotRunning)

	// Add timeout to prevent hanging
	timeout := time.After(20 * time.Second)

	select {
	case status := <-statusCh:
		fmt.Println("Container exited with status code:", status.StatusCode)
		if status.StatusCode != 0 {
			fmt.Fprintf(os.Stderr, "Container exited with non-zero status: %d", status.StatusCode)
		}
	case err := <-errCh:
		panic(err)
	case <-timeout:
		fmt.Fprintln(os.Stderr, "Container operation timed out, forcing removal")
		// Force remove the container if it's hanging
		docker.ContainerKill(ctx, tempContainer.ID, "KILL")
	}

	// // 8. Remove container (this should be automatic due to AutoRemove, but just in case)
	// err = docker.ContainerRemove(ctx, tempContainer.ID, client.ContainerRemoveOptions{RemoveVolumes: false, Force: true})
	// if err != nil {
	// 	// Don't panic here as AutoRemove might have already removed it
	// 	log.Printf("Warning: could not remove container: %v", err)
	// }

	fmt.Printf("Restored data to volume %s completed\n", targetVolume.Name)
	return nil
}
