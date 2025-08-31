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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func EncryptFile(targetFile string, publicKey string, overwrite bool) error {
	var encryptedFile string = targetFile + ".enc"
	var cmd = exec.Command("sops", "--encrypt", "--age", publicKey, targetFile)

	if _, err := os.Stat(encryptedFile); err == nil && !overwrite {
		return fmt.Errorf("an encrypted file already exists, specify -o to overwrite it")
	}

	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf("unable to encrypt file: %v", err)
	} else {

		// 1. Create the file.
		file, err := os.Create(encryptedFile)
		if err != nil {
			return fmt.Errorf("failed to create file for %s: %v", encryptedFile, err)
		}

		// 2. Use `defer` to ensure the file is closed.
		defer file.Close()

		// 3. Write the content to the file.
		_, err = file.Write(out)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}

		return fmt.Errorf("file %s encrypted successfully", encryptedFile)
	}
}

func GetPublicKeyFromDefaultLocation() (string, error) {
	keysPath, err := GetSopsAgeKeyPath()
	if err != nil {
		return "", fmt.Errorf("error getting sops age key path: %v", err)
	}
	return extractPublicKey(keysPath)
}

func extractPublicKey(path string) (string, error) {
	if filepath.Ext(path) != ".txt" {
		return "", fmt.Errorf("the provided file is not a txt file")
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("unable to locate the age file: %w", err)
		}
		return "", fmt.Errorf("error opening age file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# public key:") {
			pub := strings.TrimSpace(strings.TrimPrefix(line, "# public key:"))
			if pub == "" {
				return "", fmt.Errorf("public key line found but empty") // edge case
			}
			return pub, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading age file: %w", err)
	}

	// case 2: file exists but no public key line
	return "", fmt.Errorf("public key not found in file")
}
