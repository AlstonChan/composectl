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

// This command unsets a configuration set by
// 'composectl set' to reset the configuration
var unSetCmd = &cobra.Command{
	Use:   "unset ...",
	Short: "Unset the configuration for the application",
	Long: `Unset the configuration for the CLI application so
that it will use the default option the next time you 
execute a command.`,
	Example:   setConfigExample("unset", true),
	ValidArgs: allConfigKey,
	Args:      cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		var errorString string = ""
		for _, argument := range args {
			if errorString != "" {
				break
			}

			var hasMatchingConfig bool = false
			for _, configKey := range allConfigKey {
				if argument == configKey {
					services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

					viper.Set(configKey, "")
					if err := viper.WriteConfig(); err != nil {
						// If config file doesn’t exist, create it
						if _, ok := err.(viper.ConfigFileNotFoundError); ok {
							viper.SafeWriteConfig()
						}
						errorString = err.Error()
					}

					fmt.Printf("%s unset\n", configKey)
					hasMatchingConfig = true
					break
				}
			}

			if !hasMatchingConfig {
				errorString = "Configuration not recognized: " + argument
			}
		}
		fmt.Fprintln(os.Stderr, errorString)
	},
}

func init() {
	RootCmd.AddCommand(unSetCmd)
}
