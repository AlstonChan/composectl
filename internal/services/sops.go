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
	"runtime"

	"github.com/AlstonChan/composectl/internal/config"
)

func GetSopsAgeKeyPath() (string, error) {
	keysPath := os.Getenv(config.SopsAgeKeyFileEnv)
	if keysPath != "" {
		_, err := os.Stat(keysPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("age key file does not found from the path specified by %s\n", config.SopsAgeKeyFileEnv)
			} else {
				fmt.Printf("Unable to extract public key from the path specified by %s\n", config.SopsAgeKeyFileEnv)
			}
		} else {
			return keysPath, nil
		}
	}

	switch runtime.GOOS {
	case "linux":
		// Check XDG_CONFIG_HOME as the first priority
		configPath := os.Getenv("XDG_CONFIG_HOME")
		if configPath != "" {
			// Return the path if it exists
			var keysPath string = filepath.Join(configPath, "sops", "age", "keys.txt")
			if _, err := os.Stat(keysPath); err == nil {
				return keysPath, nil
			}
		}

		// Fallback to $HOME/.config/sops/age/keys.txt to check for keys.txt
		userDir, error := os.UserHomeDir()
		if error != nil {
			return "", fmt.Errorf("unable to determine $HOME directory to locate keys.txt")
		}

		// Return the path if it exists
		var keysPath string = filepath.Join(userDir, ".config", "sops", "age", "keys.txt")
		if _, err := os.Stat(keysPath); err == nil {
			return keysPath, nil
		} else if configPath != "" {
			// If XDG_CONFIG_HOME is set but the keys.txt file does not exist, return an error
			return "", fmt.Errorf("XDG_CONFIG_HOME is set, but the keys.txt file does not exist at %s", keysPath)
		} else {
			// If neither XDG_CONFIG_HOME nor $HOME/.config/sops/age/keys.txt exists, return an error
			return "", fmt.Errorf("the keys.txt file does not exist at %s", keysPath)
		}
	case "windows":
		appDataPath := os.Getenv("APPDATA")
		if appDataPath != "" {
			// Return the path if it exists
			var keysPath string = filepath.Join(appDataPath, "sops", "age", "keys.txt")
			if _, err := os.Stat(keysPath); err == nil {
				return keysPath, nil
			}
		}

		// Fallback to $HOME/.config/sops/age/keys.txt
		userDir, error := os.UserHomeDir()
		if error != nil {
			return "", fmt.Errorf("unable to determine $HOME directory to locate keys.txt")
		}

		var keysPath string = filepath.Join(userDir, ".config", "sops", "age", "keys.txt")
		if _, err := os.Stat(keysPath); err == nil {
			return keysPath, nil
		} else {
			return "", fmt.Errorf("unable to determine the path to keys.txt")
		}
	}

	return "", fmt.Errorf("no key file found")
}
