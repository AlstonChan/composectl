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
	"time"

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
			fmt.Fprintln(os.Stderr, "Either the service name or sequence must be specified correctly!")
			return
		}

		if path == "" && remote == "" {
			fmt.Fprintln(os.Stderr, "Either the backup file path or remote location must be specified!")
			return
		}
		if path != "" && remote != "" {
			fmt.Fprintln(os.Stderr, "Cannot use both path and remote to restore backup")
			return
		}

		if dayOffset <= 0 {
			fmt.Fprintln(os.Stderr, "day offset must be a positive number")
			return
		}
		var dateToRestoreAfter = time.Now().AddDate(0, 0, -dayOffset+1)

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
			fileData, err := os.ReadFile(fullBackupPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to read the file at %s: %v\n", fullBackupPath, err)
				return
			}

			if filepath.Ext(fullBackupPath) == ".gpg" {
				bytePassword, err := services.PromptPassphrase()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				// GPG decrypt the content
				decryptedContent, err := services.GpgDecryptFile(fileData, bytePassword)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				fileData = decryptedContent.Bytes()
			}

			err = services.RestoreAllDockerVolume(dockerClient, ctx, fileData)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		} else if remote != "" {
			switch remote {
			case "s3":
				s3Client, err := services.GetAwsAccount(ctx)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				s3Bucket, err := services.GetS3BackupStoreBucket(ctx, s3Client, CONFIG_AWS_S3_BUCKET)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				bucketExists, err := services.ValidateS3BucketExists(ctx, s3Client, s3Bucket, name)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				if !bucketExists {
					fmt.Fprintf(os.Stderr, "service directory not found: %v\n", name)
					return
				}

				backupFilename, err := services.GetFileFromBucket(ctx, s3Client, s3Bucket, name, dateToRestoreAfter)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				fileData, err := services.S3DownloadToMemory(ctx, s3Client, s3Bucket, backupFilename)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				if filepath.Ext(backupFilename) == ".gpg" {
					bytePassword, err := services.PromptPassphrase()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						return
					}

					// GPG decrypt the content
					decryptedContent, err := services.GpgDecryptFile(fileData, bytePassword)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						return
					}

					fileData = decryptedContent.Bytes()
				}

				err = services.RestoreAllDockerVolume(dockerClient, ctx, fileData)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			case "azure":
				fmt.Fprintln(os.Stderr, "Not implemented yet")
			}
		} else {
			fmt.Fprintln(os.Stderr, "Internal application error, unknown decision tree")
			return
		}
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
