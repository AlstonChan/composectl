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
	"strings"
	"unicode/utf8"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// The default path to the SelfHostCompose repository
	CONFIG_REPO_PATH = "repo-path"
	// The default age publick key to use for encryption
	CONFIG_AGE_PUBKEY = "age-pubkey"
	// The default aws s3 bucket to restore the backup from
	CONFIG_AWS_S3_BUCKET = "s3-bucket"
)

var allConfigKey = []string{
	CONFIG_REPO_PATH,
	CONFIG_AGE_PUBKEY,
	CONFIG_AWS_S3_BUCKET,
}

// Set the configuration for the composectl application, so
// that you can avoid specifying a flag everytime a relavant
// command is run that uses the config
var setCmd = &cobra.Command{
	Use:   "set ...",
	Short: "Set the configuration for the application",
	Long: strings.ReplaceAll(
		`Set the configuration for the CLI application so
that it will remember the next time you execute a 
command.

It will create a .composectl directory which SHOULD be 
git ignored. 

The .composectl directory will be created besides the
executable unless the CONFIG_DIR_ENV env is set to
the path of the .composectl directory`,
		"CONFIG_DIR_ENV", config.ConfigDirEnv),
	Example:   setConfigExample("set", false),
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

			parts := strings.SplitN(argument, "=", 2) // Split into key and value
			if len(parts) != 2 {
				errorString = "invalid format, expected key=value"
				break
			}
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))

			key := parts[0]
			value := parts[1]

			switch {
			case strings.HasPrefix(argument, CONFIG_REPO_PATH):
				// Resolve to absolute path
				absPath, err := filepath.Abs(value)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
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
			case strings.HasPrefix(argument, CONFIG_AGE_PUBKEY):
				if utf8.RuneCountInString(value) != 62 {
					fmt.Printf("The public key provided is invalid")
					continue
				}

				viper.Set(key, value)
				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}
				fmt.Printf("Age public key set to %s\n", value)
			case strings.HasPrefix(argument, CONFIG_AWS_S3_BUCKET):
				viper.Set(key, value)
				if err := viper.WriteConfig(); err != nil {
					// If config file doesn’t exist, create it
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						viper.SafeWriteConfig()
					}
					errorString = err.Error()
				}
				fmt.Printf("AWS default restoration s3 bucket set to %s\n", value)
			default:
				errorString = "Configuration not recognized: " + argument
			}
		}

		if errorString != "" {
			fmt.Fprintln(os.Stderr, errorString)
		}
	},
}

func setConfigExample(command string, noValue bool) string {
	if noValue {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(`$ composectl COMMAND repo-path
$ composectl COMMAND age-pubkey`, "repo-path", CONFIG_REPO_PATH), "age-pubkey", CONFIG_AGE_PUBKEY), "COMMAND", command)
	}

	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(`$ composectl COMMAND repo-path=./
$ composectl COMMAND age-pubkey=age1...`, "repo-path", CONFIG_REPO_PATH), "age-pubkey", CONFIG_AGE_PUBKEY), "COMMAND", command)
}

func init() {
	RootCmd.AddCommand(setCmd)
}
