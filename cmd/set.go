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
	"strings"
	"unicode/utf8"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set ...",
	Short: "Set the configuration of the application",
	Long: strings.ReplaceAll(
		`Set the configuration of the CLI application so
that it will remember the next time you execute a 
command.

It will create a .composectl directory which SHOULD be 
git ignored. 

The .composectl directory will be created besides the
executable unless the CONFIG_DIR_ENV env is set to
the path of the .composectl directory`,
		"CONFIG_DIR_ENV", config.ConfigDirEnv),
	Example: `$ composectl set repo-path=./
$ composectl set age-pubkey=age1q...`,
	ValidArgs: []string{
		"repo-path",
		"age-pubkey",
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
			case strings.HasPrefix(argument, "repo-path"):
				parts := strings.SplitN(argument, "=", 2) // Split into key and value
				if len(parts) != 2 {
					errorString = "invalid format, expected key=value"
					break
				}
				services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

				key := parts[0]
				value := parts[1]

				// Resolve to absolute path
				absPath, err := filepath.Abs(value)
				if err != nil {
					log.Fatal(err)
				}
				viper.Set(key, absPath)

				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}

				fmt.Printf("Repo root set to %s\n", absPath)
			case strings.HasPrefix(argument, "age-pubkey"):
				parts := strings.SplitN(argument, "=", 2) // Split into key and value
				if len(parts) != 2 {
					errorString = "invalid format, expected key=value"
					break
				}
				services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

				key := parts[0]
				value := parts[1]

				if utf8.RuneCountInString(value) != 62 {
					fmt.Printf("The public key provided is invalid")
					return
				}
				viper.Set(key, value)

				fmt.Printf("Age public key set to %s\n", value)
			default:
				errorString = "Configuration not recognized: " + argument
			}
		}

		if errorString != "" {
			fmt.Println((errorString))
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
