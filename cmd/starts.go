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

var startsCmd = &cobra.Command{
	Use:   "starts",
	Short: "Starts a interactive session for starting service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := deps.CheckSops(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if err := deps.CheckDockerDeps(0, 2); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		// Show user all service for them to select which service
		// to run. Docker status and decrypt status has to be
		// shown

		// For each service, decrypt the all the encrypted secrets

		// Show a list of decrypted secret file, allow user to
		// - preview: shows a maximum of 20 lines
		// - view all: spawn `less` to show the content
		// - edit: use an external editor. Search from $VISUAL,
		// then $EDITOR, then fallback to nano or vi

		// If user choose to edit, after the edit is finished.
		// Prompt the user to either proceed or encrypt before
		// proceed. Then show the list of decrypted secret file
		// again.

		// User can then procced to the next service.

		// Before actually proceeding. Scan the docker compose file
		// for external volume or network. If there are external
		// resource, check if it has been created. If not, create
		// it for the user.

		// After all service has been looped through. Ask user
		// whether to start the service now.
	},
}

func init() {
	rootCmd.AddCommand(startsCmd)
}
