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
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Show the details of the specified service",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")
		includeAllFiles, _ := cmd.Flags().GetBool("all")

		if name == "" && sequence <= 0 {
			fmt.Println("Either the service name or sequence must be specified correctly!")
			return
		}

		if err := deps.CheckDockerDeps(0, 2); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			log.Fatalf("Error resolving repo root: %v", err)
		}

		serviceList, err := services.ListAllService(repoRoot)
		if err != nil {
			log.Fatalf("Error listing services: %v", err)
		}

		if len(serviceList) == 0 {
			if sequence >= 1 {
				fmt.Printf("Service with sequence %d not found", sequence)
			} else if name != "" {
				fmt.Printf("Service with name %s not found", name)
			}
			return
		}

		// Check if service specified exists
		if sequence >= 1 {
			if len(serviceList) < sequence {
				fmt.Printf("Service with sequence %d not found", sequence)
				return
			}
			name = serviceList[sequence-1]
		} else if name != "" {
			var serviceIndex int = slices.IndexFunc(serviceList, func(s string) bool { return s == name })
			if serviceIndex == -1 {
				fmt.Printf("Service with name \"%s\" not found", name)
				return
			}
			sequence = serviceIndex + 1
		}

		// Main Info
		fmt.Println("================================")
		fmt.Printf("Docker service %s(%d)\n", name, sequence)
		fmt.Println("================================")

		decryptStatus := services.GetDecryptedFilesStatus(repoRoot, name)

		var serviceDirectory string = filepath.Join(repoRoot, config.DockerServicesDir, name)
		state, err := services.GetServiceState(serviceDirectory)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Docker status: %s\n", services.GetServiceStatusString(state))
		fmt.Printf("Decryption status: %s\n\n", services.GetDecryptedStatusString(decryptStatus))

		// List files
		files, err := services.ResolveServiceDetails(repoRoot, name, !includeAllFiles)
		if err != nil {
			log.Fatalf("Error resolving service's details: %v", err)
		}

		if includeAllFiles {
			fmt.Printf("All %s service files\n", name)
		} else {
			fmt.Printf("All %s service secrets\n", name)
		}
		for _, file := range files {
			fmt.Printf("- %s", file.Filename)
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
