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

	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/spf13/cobra"
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

		if err := deps.CheckSops(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Printf("decrypt: name - %s, sequence - %d, decrypt-all - %t, index - %d", name, sequence, decryptAll, index)
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringP("name", "n", "", "The name of the service")
	decryptCmd.Flags().IntP("sequence", "s", 0, "The sequence of the service")
	decryptCmd.Flags().BoolP("decrypt-all", "a", false, "Decrypt all secrets of the service")
	decryptCmd.Flags().IntP("index", "i", 0, "Specify the index of the secrets to decrypt")
}
