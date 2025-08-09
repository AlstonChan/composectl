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
	"path/filepath"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/docker"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the selfhost repo with status",
	Run: func(cmd *cobra.Command, args []string) {
		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			log.Fatalf("Error resolving repo root: %v", err)
		}

		serviceList, err := services.ListAllService(repoRoot)
		if err != nil {
			log.Fatalf("Error listing services: %v", err)
		}

		fmt.Printf("Repo root: %s\n\n", repoRoot)
		if len(serviceList) == 0 {
			fmt.Println("No services found.")
			return
		}

		for i, s := range serviceList {
			var sequence int = i + 1
			var serviceDirectory string = filepath.Join(repoRoot, config.DockerServicesDir, s)
			var serviceStatus string = "Unused"

			decrypted := services.IsEnvDecrypted(fmt.Sprintf("%s/.env", serviceDirectory))

			state, err := docker.GetServiceState(serviceDirectory)
			if err != nil {
				log.Fatal(err)
			}

			switch state {
			case docker.Running:
				serviceStatus = "Running"
			case docker.PartiallyRunning:
				serviceStatus = "Partial"
			case docker.Stopped:
				serviceStatus = "Stopped"
			}

			fmt.Printf("%2d - %-25s  Status: %-10s  Decrypted: %-5t\n", sequence, s, serviceStatus, decrypted)
		}
		fmt.Print("\n")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
