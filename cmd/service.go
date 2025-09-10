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
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Show the details of the specified service",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")
		includeAllFiles, _ := cmd.Flags().GetBool("all")

		if name == "" && sequence <= 0 {
			fmt.Fprintln(os.Stderr, "Either the service name or sequence must be specified correctly!")
			return
		}

		if err := deps.CheckDockerDeps(0, 2); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
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

		// Main Info
		fmt.Println("================================")
		fmt.Printf("Docker service %s(%d)\n", name, sequence)
		fmt.Println("================================")

		decryptStatus := services.GetDecryptedFilesStatus(repoRoot, name)

		var serviceDirectory string = filepath.Join(repoRoot, config.DockerServicesDir, name)
		states, err := services.GetAllServiceState(serviceDirectory)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}

		fmt.Printf("Decryption status: %s\n\n", services.GetDecryptedStatusString(decryptStatus))
		for _, state := range states {
			fmt.Printf("%s:\n", state.File)
			fmt.Printf("Docker status: %s\n\n", services.GetServiceStatusString(state.ServiceState))
		}

		// List files
		files, err := services.ResolveServiceFiles(repoRoot, name, !includeAllFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving service's details: %v\n", err)
		}

		if includeAllFiles {
			fmt.Printf("All %s service files\n", name)
		} else {
			fmt.Printf("All %s service secrets\n", name)
		}

		var secretCount int = 1
		for _, file := range files {
			if file.IsSecrets {
				fmt.Printf("%2d. %s", secretCount, file.Filename)
				secretCount++
			} else {
				fmt.Printf("--- %s", file.Filename)
			}

			if file.HasDecryptedVersion {
				fmt.Printf(" (has decrypted version)")
			}
			fmt.Print("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().StringP("name", "n", "", "The name of the service")
	serviceCmd.Flags().IntP("sequence", "s", 0,
		"The sequence of the service. This args has precedence over the name args when both are specified")
	serviceCmd.Flags().BoolP("all", "a", false, "List all file, including non-secrets")
}
