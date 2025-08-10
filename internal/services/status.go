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
	"log"
	"os"
	"strings"
)

type ServiceDecryptionStatus int

const (
	All ServiceDecryptionStatus = iota
	Partial
	None
	NIL
)

func IsEnvDecrypted(envPath string) bool {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return false // missing file means not decrypted
	}

	// crude check: SOPS-encrypted env usually starts with "ENC[" or has sops metadata
	if strings.Contains(string(data), "sops") || strings.Contains(string(data), "ENC[") {
		return false
	}
	return true
}

func GetDecryptedFilesStatus(root string, serviceName string) ServiceDecryptionStatus {
	files, err := ResolveServiceDetails(root, serviceName, true)
	if err != nil {
		log.Fatalf("Error resolving service's details: %v", err)
	}

	var decryptedFileCount int = 0
	for _, file := range files {
		if file.HasDecryptedVersion {
			decryptedFileCount++
		}
	}

	switch {
	case len(files) == 0:
		return NIL
	case decryptedFileCount > 1 && decryptedFileCount == len(files):
		return All
	case decryptedFileCount > 1 && decryptedFileCount < len(files):
		return Partial
	case decryptedFileCount == 0:
		return None
	}

	return None
}

func GetDecryptedStatusString(status ServiceDecryptionStatus) string {
	switch status {
	case All:
		return "All"
	case Partial:
		return "Partial"
	case None:
		return "None"
	case NIL:
		return "NIL"
	default:
		return fmt.Sprintf("ServiceDecryptionStatus(%d)", status)
	}
}

// IsEncryptedFile checks if a file path ends with ".enc" or contains ".enc" anywhere in its name.
// It returns true if either condition is met, otherwise false.
func IsEncryptedFile(filePath string) bool {
	// Check if the path ends with ".enc"
	if strings.HasSuffix(filePath, ".enc") {
		return true
	}

	// Check if the path contains ".enc"
	if strings.Contains(filePath, ".enc.") {
		return true
	}

	// If neither condition is met, return false
	return false
}
