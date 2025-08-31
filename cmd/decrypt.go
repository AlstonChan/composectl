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

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt the secrets of the specified service",
	Long: `Decrypt one or all secrets of the specified service.
	
	If the secrets to decrypt already has a same filename existed, 
	it will overwrite the file content.
	
	To get the index of the secret that you want to decrypt. Use 
	the service command`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		sequence, _ := cmd.Flags().GetInt("sequence")

		decryptAll, _ := cmd.Flags().GetBool("decrypt-all")
		index, _ := cmd.Flags().GetInt("index")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		if name == "" && sequence <= 0 {
			fmt.Println("Either the service name or sequence must be specified correctly!")
			return
		}

		if !decryptAll && index <= 0 {
			fmt.Println("You have to specify an index to decrypt or use \"-a\" to decrypt all secrets")
			return
		}

		if err := deps.CheckSops(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
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
		}

		serviceLists, err := services.ValidateService(repoRoot, &sequence, &name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		// Service does not exists
		if serviceLists == nil && err == nil {
			return
		}

		if decryptAll {
			err = services.DecryptAllFile(repoRoot, name, overwrite)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to decrypt file for service %s: %v\n", name, err)
				return
			}
		} else {
			files, err := services.ResolveServiceFiles(repoRoot, name, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving service's details: %v\n", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("this service does not have any file to decrypt")
				return
			}

			if index < 0 || index > len(files) || files[index-1] == (services.ServiceFile{}) {
				fmt.Fprintf(os.Stderr, "the file given index %d does not exists\n", index)
				return
			}

			err = services.DecryptFile(repoRoot, name, index, files[index-1], overwrite)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to decrypt file for service %s: %v\n", name, err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringP("name", "n", "", "The name of the service")
	decryptCmd.Flags().IntP("sequence", "s", 0, "The sequence of the service")
	decryptCmd.Flags().BoolP("decrypt-all", "a", false, "Decrypt all secrets of the service")
	decryptCmd.Flags().IntP("index", "i", 0, "Specify the index of the secrets to decrypt")
	decryptCmd.Flags().BoolP("overwrite", "o", false, "Whether to overwrite the file if it already exists")
}
