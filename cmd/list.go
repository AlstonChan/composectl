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

	"github.com/AlstonChan/composectl/internal/constants"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
)

var repoPath string // holds --repo-path flag value

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the selfhost repo with status",
	Run: func(cmd *cobra.Command, args []string) {
		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			log.Fatalf("Error resolving repo root: %v", err)
		}

		serviceList, err := services.ListAll(repoRoot)
		if err != nil {
			log.Fatalf("Error listing services: %v", err)
		}

		fmt.Printf("Repo root: %s\n", repoRoot)
		if len(serviceList) == 0 {
			fmt.Println("No services found.")
			return
		}

		for _, s := range serviceList {
			running := services.IsServiceRunning(s)
			decrypted := services.IsEnvDecrypted(fmt.Sprintf("%s/%s/%s/.env", repoRoot, constants.DockerServicesDir, s))
			fmt.Printf(" - %-25s  Running: %-5t  Decrypted: %-5t\n", s, running, decrypted)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&repoPath, "repo-path", "", "Path to selfhost repo (overrides binary location)")
}
