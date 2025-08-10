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
	"os"
	"path/filepath"
	"sort"

	"github.com/AlstonChan/composectl/internal/config"
)

func ListAllService(root string) ([]string, error) {
	var servicesPath = filepath.Join(root, config.DockerServicesDir)
	var services []string

	entries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, err
	}

	// Ensure that the directory are already sorted to have
	// consistent sequence for the service
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			composePath := filepath.Join(servicesPath, entry.Name(), "compose.yml")
			if fileExists(composePath) {
				services = append(services, entry.Name())
			}
		}
	}
	return services, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
