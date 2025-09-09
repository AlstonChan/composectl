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
	"strings"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unSetCmd = &cobra.Command{
	Use:   "unset ...",
	Short: "Unset the configuration for the application",
	Long: `Unset the configuration for the CLI application so
that it will use the default option the next time you 
execute a command.`,
	Example: setConfigExample("unset", true),
	ValidArgs: []string{
		CONFIG_REPO_PATH,
		CONFIG_AGE_PUBKEY,
	},
	Args: cobra.MinimumNArgs(1),
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

			switch {
			case strings.HasPrefix(argument, CONFIG_REPO_PATH):
				services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

				viper.Set(CONFIG_REPO_PATH, "")
				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}

				fmt.Printf("Repository path unset\n")
			case strings.HasPrefix(argument, CONFIG_AGE_PUBKEY):
				services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

				viper.Set(CONFIG_AGE_PUBKEY, "")
				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}

				fmt.Printf("Age public key unset\n")
			case strings.HasPrefix(argument, CONFIG_AWS_S3_BUCKET):
				services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

				viper.Set(CONFIG_AWS_S3_BUCKET, "")
				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}

				fmt.Printf("WS default restoration s3 bucket unset\n")
			default:
				errorString = "Configuration not recognized: " + argument
			}
		}

		if errorString != "" {
			fmt.Fprintln(os.Stderr, errorString)
		}
	},
}

func init() {
	rootCmd.AddCommand(unSetCmd)
}
