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

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// This command encrypts a file into secrets given the file path
// that is relative to the service, so that you do not have to
// write a long path.
// This command will automatically detect the file type based
// on the target file extension, and use the correct format
// for encryption. As for file format that isn't recognize, it
// will be encrypted as a json file, but decryption with the
// 'composectl decrypt' would recognize it and still decrypt
// it back to the original file
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt the secrets of the specified service",
	Example: `  Encrypt a docker service secrets:

  # by service sequence (as per 'composectl list') for a file relative to the service root
  composectl encrypt -s 12 -f .env

  # by service name (as per 'composectl list') for a file relative to the service root
  composectl encrypt -n gitea -f .app.ini

  # to overwrite existing encrypted secret
  composectl encrypt -n gitea -f private-key.pem -o

  # for all secrets of the service
  composectl decrypt -n gitea -a

  # to specify a age public key if not set with 'composectl set'
  composectl encrypt -n gitea -f config.yaml -p age1....
`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")

		publicKey, _ := cmd.Flags().GetString("pubkey")
		file, _ := cmd.Flags().GetString("file")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		if name == "" && sequence <= 0 {
			fmt.Fprintln(os.Stderr, "Either the service name or sequence must be specified correctly!")
			return
		}

		if err := deps.CheckSops(); err != nil {
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

		if publicKey == "" {
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
			if val := viper.GetString(CONFIG_AGE_PUBKEY); val != "" {
				publicKey = val
			} else {
				var err error
				publicKey, err = services.GetPublicKeyFromDefaultLocation()
				if err != nil {
					fmt.Fprintf(os.Stderr, "An error occurred while getting the public key: %v\n", err)
					return
				}
			}
		}

		var targetFile string = filepath.Join(repoRoot, config.DockerServicesDir, name, file)
		if _, err := os.Stat(targetFile); err != nil {
			fmt.Fprintf(os.Stderr, "The provided file %s does not exists!\n", targetFile)
			return
		}

		if err := services.EncryptFile(targetFile, publicKey, overwrite); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringP("name", "n", "", "The name of the service")
	encryptCmd.Flags().IntP("sequence", "s", 0, "The sequence of the service")
	encryptCmd.Flags().StringP("file", "f", "", "The filename/file path to encrypt of the service")
	encryptCmd.Flags().StringP("pubkey", "p", "", "The age public key to encrypt secrets")
	encryptCmd.Flags().BoolP("overwrite", "o", false, "Whether to overwrite the file if it already exists")
}
