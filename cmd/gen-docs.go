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

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// This command generates all documentation in markdown format into
// the specified path if specified, that defaults to ./docs directory
var genDocsCmd = &cobra.Command{
	Use:   "gen-docs",
	Short: "Generate documentation of the application in markdown format",
	Example: `  To generate the documentation to a directory:

  # to default 'docs' directory
  composectl gen-docs

  # to a specific directory
  composectl gen-docs -p ../my-docs`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		fullPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get the full path of the directory: %v\n", err)
			return
		}

		err = os.MkdirAll(fullPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory: %v\n", err)
			return
		}

		err = doc.GenMarkdownTree(RootCmd, fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate documentation: %v\n", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(genDocsCmd)
	genDocsCmd.Flags().StringP("path", "p", "docs", "The path of the docs to be generated")
}
