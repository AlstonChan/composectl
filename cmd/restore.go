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

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/moby/moby/client"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the service's data from backup",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")

		path, _ := cmd.Flags().GetString("path")
		remote, _ := cmd.Flags().GetString("remote")

		dayOffset, _ := cmd.Flags().GetInt("day")

		if name == "" && sequence <= 0 {
			fmt.Println("Either the service name or sequence must be specified correctly!")
			return
		}

		if path == "" && remote == "" {
			fmt.Println("Either the backup file path or remote location must be specified!")
			return
		}
		if path != "" && remote != "" {
			fmt.Println("Cannot use both path and remote to restore backup")
			return
		}

		if dayOffset <= 0 {
			fmt.Println("day offset must be a positive number")
			return
		}

		if err := deps.CheckDockerDeps(0, 2); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return
		}

		dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to create docker client: ", err)
			return
		}

		if repoPath == "" {
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
			if val := viper.GetString(CONFIG_REPO_PATH); val != "" {
				repoPath = val
			}
		}

		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving repo root: %v\n", err)
			return
		}

		serviceLists, err := services.ValidateService(repoRoot, &sequence, &name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		if serviceLists == nil && err == nil {
			return
		}

		ctx := context.Background()
		if path != "" {
			// Get the complete path to the backup file
			fullBackupPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to parse the full path to the backup file")
				return
			}

			// Check if the backup file exists
			if _, err := os.Stat(fullBackupPath); err != nil {
				fmt.Fprintln(os.Stderr, "The file does not exists or program has no permission")
				return
			}

			// Read encrypted file
			encBytes, err := os.ReadFile(fullBackupPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to read the file at %s: %v\n", fullBackupPath, err)
				return
			}

			bytePassword, err := services.PromptPassphrase()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			// GPG decrypt the content
			decryptedContent, err := services.GpgDecryptFile(encBytes, bytePassword)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			err = services.RestoreAllDockerVolume(dockerClient, ctx, decryptedContent.Bytes())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		} else if remote != "" {
			// Handle restoring backup from remote location
		} else {
			fmt.Fprintln(os.Stderr, "Internal application error, unknown decision tree")
			return
		}
		fmt.Println("Command not implemented yet...; service name:", name, "sequence:", sequence)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringP("name", "n", "", "The name of the service")
	restoreCmd.Flags().IntP("sequence", "s", 0,
		"The sequence of the service. This args has precedence over the name args when both are specified")
	restoreCmd.Flags().StringP("path", "p", "", "The path to the local backup file (mutually exclusive with --remote)")
	restoreCmd.Flags().String("remote", "", "The remote location to restore the backup from (mutually exclusive with --path)")
	restoreCmd.Flags().IntP("day", "d", 1, "Relative day offset for backup (1 = latest, 2 = yesterday, etc.)")
}
