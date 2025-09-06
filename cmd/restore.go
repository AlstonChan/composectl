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

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the service's data from backup",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")

		if name == "" && sequence <= 0 {
			fmt.Println("Either the service name or sequence must be specified correctly!")
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

		fmt.Println("Command not implemented yet...; service name:", name, "sequence:", sequence)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringP("name", "n", "", "The name of the service")
	restoreCmd.Flags().IntP("sequence", "s", 0,
		"The sequence of the service. This args has precedence over the name args when both are specified")
}
