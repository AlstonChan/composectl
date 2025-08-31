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
	"strings"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/spf13/viper"
)

// Priority: CLI flag > env var > binary dir
func ResolveRepoRoot(cliPath string) (string, error) {
	if cliPath != "" {
		return verifyRepoRoot(cliPath)
	}

	if envPath := os.Getenv(config.RepoPathEnv); envPath != "" {
		return verifyRepoRoot(envPath)
	}

	exeDir, err := GetExecutableDir()
	if err != nil {
		return "", err
	}
	return verifyRepoRoot(exeDir)
}

func GetExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func verifyRepoRoot(path string) (string, error) {
	var gitDir = filepath.Join(path, ".git")
	var dockerServicesDir = filepath.Join(path, config.DockerServicesDir)

	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("not a valid selfhost repo root at %s (missing .git)", path)
	}

	if _, err := os.Stat(dockerServicesDir); err != nil {
		return "", fmt.Errorf("not a valid selfhost repo root at %s (missing docker_services)", path)
	}
	return path, nil
}

func CreateLocalCacheDir(path string) (string, error) {
	// use default path if empty
	if path == "" {
		var exePath, err = GetExecutableDir()
		if err != nil {
			return "", err
		}
		path = exePath
	}

	if !strings.HasSuffix(path, config.LocalConfigDir) {
		path = filepath.Join(path, config.LocalConfigDir)
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		} else {
			return "", err
		}
	}

	viper.AddConfigPath(path)
	initConfig()

	return path, nil
}

func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	// Create file if missing
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil && !os.IsExist(err) {
				fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
}
