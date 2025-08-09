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
)

// Priority: CLI flag > env var > binary dir
func ResolveRepoRoot(cliPath string) (string, error) {
	if cliPath != "" {
		return verifyRepoRoot(cliPath)
	}

	if envPath := os.Getenv("MYCLI_REPO_PATH"); envPath != "" {
		return verifyRepoRoot(envPath)
	}

	exeDir, err := getExecutableDir()
	if err != nil {
		return "", err
	}
	return verifyRepoRoot(exeDir)
}

func getExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func verifyRepoRoot(path string) (string, error) {
	var gitDir = filepath.Join(path, ".git")
	var dockerServicesDir = filepath.Join(path, "docker_services")

	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("not a valid selfhost repo root at %s (missing .git)", path)
	}

	if _, err := os.Stat(dockerServicesDir); err != nil {
		return "", fmt.Errorf("not a valid selfhost repo root at %s (missing docker_services)", path)
	}
	return path, nil
}
