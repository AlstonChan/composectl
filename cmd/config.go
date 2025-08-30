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
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the configuration that has been set for the application",
	Run: func(cmd *cobra.Command, args []string) {
		services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

		var repoPath string = viper.GetString(CONFIG_REPO_PATH)
		var agePubKey string = viper.GetString(CONFIG_AGE_PUBKEY)

		fmt.Println("composectl configuration")
		fmt.Printf("Repository path: %s\n", orDefault(repoPath, "Not set"))
		fmt.Printf("Age public key: %s\n", orDefault(agePubKey, "Not set"))
	},
}

func orDefault(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func init() {
	rootCmd.AddCommand(configCmd)
}
