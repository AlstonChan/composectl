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

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt the secrets of the specified service",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")

		publicKey, _ := cmd.Flags().GetString("pubkey")
		file, _ := cmd.Flags().GetString("file")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		if name == "" && sequence <= 0 {
			fmt.Println("Either the service name or sequence must be specified correctly!")
			return
		}

		if err := deps.CheckSops(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if repoPath == "" {
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
			if val := viper.GetString("repo-path"); val != "" {
				repoPath = val
			}
		}

		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			log.Fatalf("Error resolving repo root: %v", err)
		}

		serviceLists, err := services.ValidateService(repoRoot, &sequence, &name)
		if err != nil {
			fmt.Println(err.Error())
		}

		if serviceLists == nil && err == nil {
			return
		}

		if publicKey == "" {
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
			if val := viper.GetString("age-pubkey"); val != "" {
				publicKey = val
			} else {
				var err error
				publicKey, err = services.GetPublicKeyFromDefaultLocation()
				if err != nil {
					fmt.Printf("An error occurred while getting the public key: %v", err)
					return
				}
			}
		}

		var targetFile string = filepath.Join(repoRoot, config.DockerServicesDir, name, file)
		if _, err := os.Stat(targetFile); err != nil {
			fmt.Printf("The provided file %s does not exists!", targetFile)
			return
		}

		if err := services.EncryptFile(targetFile, publicKey, overwrite); err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringP("name", "n", "", "The name of the service")
	encryptCmd.Flags().IntP("sequence", "s", 0, "The sequence of the service")
	encryptCmd.Flags().StringP("file", "f", "", "The filename/file path to encrypt of the service")
	encryptCmd.Flags().StringP("pubkey", "p", "", "The age public key to encrypt secrets")
	encryptCmd.Flags().BoolP("overwrite", "o", false, "Whether to overwrite the file if it already exists")
}
