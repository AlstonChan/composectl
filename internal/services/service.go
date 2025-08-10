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

package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlstonChan/composectl/internal/config"
)

type ServiceFile struct {
	Filename            string
	HasDecryptedVersion bool
}

func ResolveServiceDetails(root string, serviceName string, secretsOnly bool) ([]ServiceFile, error) {
	var servicePath = filepath.Join(root, config.DockerServicesDir, serviceName)

	// Recursively search the directory for secrets
	var files []ServiceFile

	var err error = filepath.Walk(servicePath, func(path string, info os.FileInfo, err error) error {
		// If there was an error accessing the path, return it.
		if err != nil {
			return err
		}

		// Skip the root path itself, as we only want to list its contents.
		// We don't need directory too
		if path == servicePath || info.IsDir() {
			return nil
		}

		var hasDecryptedVersion bool = false
		if IsEncryptedFile(path) {
			var decryptedFilename = strings.ReplaceAll(path, ".enc", "")
			_, err := os.Stat(decryptedFilename)
			if err == nil {
				hasDecryptedVersion = true
			}
		}

		if secretsOnly && !IsEncryptedFile(path) {
			return nil
		}

		// Get the path relative to the root.
		// For example, if rootPath is "SelfHostService/docker_services/treafik" and the file is
		// "/home/alston/SelfHostService/docker_services/traefik/env.enc",
		// this will return "traefik/env.enc".
		relativePath, err := filepath.Rel(servicePath, path)
		if err != nil {
			return fmt.Errorf("could not get relative path for %s: %w", path, err)
		}

		files = append(files, ServiceFile{Filename: relativePath, HasDecryptedVersion: hasDecryptedVersion})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", servicePath, err)
	}

	sortByDepth(files)

	return files, nil
}

// sortByDepth sorts a slice of file paths in place by the number of
// directory separators they contain. Paths with fewer separators (less
// nested) will appear first.
func sortByDepth(files []ServiceFile) {
	sort.Slice(files, func(i, j int) bool {
		// Use strings.Count to count the number of path separators.
		// filepath.Separator is a rune that represents the OS-specific path separator ('/' or '\').
		countI := strings.Count(files[i].Filename, string(filepath.Separator))
		countJ := strings.Count(files[j].Filename, string(filepath.Separator))

		// Return true if the number of separators in files[i] is less than
		// the number in files[j], which means it is less nested.
		return countI < countJ
	})
}
