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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/AlstonChan/composectl/internal/config"
)

type ServiceFile struct {
	Filename            string
	HasDecryptedVersion bool
	IsSecrets           bool
}

func ResolveServiceFiles(root string, serviceName string, secretsOnly bool) ([]ServiceFile, error) {
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
		var isEncryptedFile bool = IsEncryptedFile(path)
		if isEncryptedFile {
			var decryptedFilename = strings.ReplaceAll(path, ".enc", "")
			_, err := os.Stat(decryptedFilename)
			if err == nil {
				hasDecryptedVersion = true
			}
		}

		if secretsOnly && !isEncryptedFile {
			return nil
		}

		// Get the path relative to the root.
		// For example, if rootPath is "SelfHostService/docker_services/traefik" and the file is
		// "/home/user/SelfHostService/docker_services/traefik/env.enc",
		// this will return "traefik/env.enc".
		relativePath, err := filepath.Rel(servicePath, path)
		if err != nil {
			return fmt.Errorf("could not get relative path for %s: %w", path, err)
		}

		files = append(files, ServiceFile{Filename: relativePath, HasDecryptedVersion: hasDecryptedVersion, IsSecrets: isEncryptedFile})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", servicePath, err)
	}

	// This is very important as the index of the file will
	// be used by other command to perform action on file index
	sortByDepth(files)

	return files, nil
}

func ValidateService(repoRoot string, sequence *int, name *string) ([]string, error) {
	serviceList, err := ListAllService(repoRoot)
	if err != nil {
		log.Fatalf("Error listing services: %v", err)
	}

	if len(serviceList) == 0 {
		if *sequence >= 1 {
			return nil, fmt.Errorf("service with sequence %d not found", *sequence)
		} else if *name != "" {
			return nil, fmt.Errorf("service with name %s not found", *name)
		}
	}

	// Check if service specified exists
	if *sequence >= 1 {
		if len(serviceList) < *sequence {
			return nil, errors.New("Service with sequence " + strconv.Itoa(*sequence) + " not found")
		}
		*name = serviceList[(*sequence)-1]
	} else if *name != "" {
		var serviceIndex int = slices.IndexFunc(serviceList, func(s string) bool { return s == *name })
		if serviceIndex == -1 {
			return nil, errors.New("Service with name \"" + *name + "\" not found")
		}
		*sequence = serviceIndex + 1
	}

	return serviceList, nil
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
