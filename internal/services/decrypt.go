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
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlstonChan/composectl/internal/config"
)

func DecrypAllFile(repoRoot string, name string) error {
	files, err := ResolveServiceFiles(repoRoot, name, true)
	if err != nil {
		return fmt.Errorf("error resolving service's details: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("this service does not have any file to decrypt")
	}

	for index, file := range files {
		DecryptFile(repoRoot, name, index+1, file)
	}

	return nil
}

func DecryptFile(repoRoot string, name string, index int, file ServiceFile) error {
	var servicePath string = filepath.Join(repoRoot, config.DockerServicesDir, name)
	var targetFilePath string = filepath.Join(servicePath, file.Filename)

	_, err := os.Stat(targetFilePath)
	if err != nil {
		return fmt.Errorf("the file given index %d cannot be found at %s", index, targetFilePath)
	}

	fileType, filename := parseEncFilename(targetFilePath, file.Filename)

	actualFilePath, err := filepath.Abs(targetFilePath)
	if err != nil {
		return err
	}

	var cmd = exec.Command("sops", "--input-type", fileType, "--output-type", fileType,
		"-d", actualFilePath)

	if out, err := cmd.Output(); err != nil {
		if err.Error() == "exit status 128" {
			userDir, userDirErr := os.UserHomeDir()
			if userDirErr != nil {
				fmt.Println("Unable to access user directory")
			} else {
				fmt.Printf("Missing age key at %s/.config/sops/age/keys.txt:", userDir)
			}
		} else {
			fmt.Println("Unable to decrypt file:", err, cmd)
		}
	} else {
		decryptedFilePath, err := filepath.Abs(filepath.Join(servicePath, filename))
		if err != nil {
			return err
		}

		// 1. Create the file.
		file, err := os.Create(decryptedFilePath)
		if err != nil {
			log.Fatalf("failed to create file for %s: %v", filename, err)
		}

		// 2. Use `defer` to ensure the file is closed.
		defer file.Close()

		// 3. Write the content to the file.
		_, err = file.Write(out)
		if err != nil {
			log.Fatalf("failed to write to file: %v", err)
		}

		fmt.Printf("File %s decrypted successfully\n", filename)
	}
	return nil
}

func parseEncFilename(targetFilePath string, file string) (fileType string, decryptedFilename string) {
	switch {
	case strings.HasSuffix(targetFilePath, ".env.enc"):
		fileType = "dotenv"
		decryptedFilename = strings.TrimSuffix(file, ".enc")

	case strings.HasSuffix(targetFilePath, ".enc.yml"):
		fileType = "yaml"
		decryptedFilename = strings.TrimSuffix(file, ".enc.yml") + ".yml"
	case strings.HasSuffix(targetFilePath, ".enc.yaml"):
		fileType = "yaml"
		decryptedFilename = strings.TrimSuffix(file, ".enc.yaml") + ".yaml"
	case strings.HasSuffix(targetFilePath, ".yml.enc"):
		fileType = "yaml"
		decryptedFilename = strings.TrimSuffix(file, ".yml.enc") + ".yml"
	case strings.HasSuffix(targetFilePath, ".yaml.enc"):
		fileType = "yaml"
		decryptedFilename = strings.TrimSuffix(file, ".yaml.enc") + ".yaml"

	case strings.HasSuffix(targetFilePath, ".enc.toml"):
		fileType = "toml"
		decryptedFilename = strings.TrimSuffix(file, ".enc.toml") + ".toml"
	case strings.HasSuffix(targetFilePath, ".toml.enc"):
		fileType = "toml"
		decryptedFilename = strings.TrimSuffix(file, ".toml.enc") + ".toml"

	case strings.HasSuffix(targetFilePath, ".enc.json"):
		fileType = "json"
		decryptedFilename = strings.TrimSuffix(file, ".enc.json") + ".json"
	case strings.HasSuffix(targetFilePath, ".json.enc"):
		fileType = "json"
		decryptedFilename = strings.TrimSuffix(file, ".json.enc") + ".json"
	}

	return fileType, decryptedFilename
}
