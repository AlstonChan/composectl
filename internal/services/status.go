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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
)

type ServiceDecryptionStatus int
type ServiceState int

const (
	All ServiceDecryptionStatus = iota
	Partial
	None
	NIL
)

const (
	Stopped ServiceState = iota
	PartiallyRunning
	Running
	Unused
)

type composeService struct {
	Name   string `json:"Service"`
	Status string `json:"State"`
}

type composeFileStatus struct {
	ServiceState
	Label string
	File  string
}

func GetActiveServiceState(projectDir string) (composeFileStatus, error) {
	var services, err = GetAllServiceState(projectDir)
	if err != nil {
		return composeFileStatus{ServiceState: Unused, Label: ""}, err
	}

	var activeService composeFileStatus
	for _, service := range services {
		if service.ServiceState != PartiallyRunning || service.ServiceState == Running {
			activeService = service
			break
		}
	}

	return activeService, nil
}

func GetAllServiceState(projectDir string) ([]composeFileStatus, error) {
	var allComposeFiles, err = FindComposeFiles(projectDir)
	if err != nil {
		return nil, err
	}

	var services []composeFileStatus = make([]composeFileStatus, len(allComposeFiles))
	for index, file := range allComposeFiles {
		state, err := GetServiceState(projectDir, file)
		if err != nil {
			log.Fatal(err)
		}

		label, err := ExtractComposeVariant(file)
		if err != nil {
			fmt.Printf("Error parsing docker compose filename: %s\n", file)
			continue
		}

		services[index] = composeFileStatus{ServiceState: state, Label: label, File: file}
	}

	// Sort the service label, with empty string at the first
	sort.Slice(services, func(i, j int) bool {
		li, lj := services[i].Label, services[j].Label
		if li == "" && lj != "" {
			return true // "" comes before others
		}
		if li != "" && lj == "" {
			return false // others come after ""
		}
		return li < lj // both non-empty → normal alpha order
	})

	return services, nil
}

func GetServiceState(projectDir string, composeFile string) (ServiceState, error) {
	cmd := exec.Command("docker", "compose", "-f", composeFile, "ps", "--format", "json")
	cmd.Dir = projectDir

	out, err := cmd.Output()
	if err != nil {
		return Unused, nil
	}

	// docker compose ps returns nothing when there are
	// no containers at any status.
	if len(out) == 0 {
		return Unused, nil
	}

	var services []composeService

	// Use a decoder to read (multiple) JSON objects in sequence
	dec := json.NewDecoder(bytes.NewReader(out))
	for dec.More() {
		var container composeService
		if err := dec.Decode(&container); err != nil {
			fmt.Println("Error decoding docker compose ps (json) output:", err)
			return Stopped, err
		}
		services = append(services, container)
	}

	if len(services) == 0 {
		return Stopped, nil
	}

	runningCount := 0
	for _, svc := range services {
		if svc.Status == "running" {
			runningCount++
		}
	}

	switch {
	case runningCount == len(services):
		return Running, nil
	case runningCount > 0:
		return PartiallyRunning, nil
	default:
		return Unused, nil
	}
}

func GetServiceStatusString(state ServiceState) string {
	switch state {
	case Stopped:
		return "Stopped"
	case PartiallyRunning:
		return "Partial"
	case Running:
		return "Running"
	case Unused:
		return "Unused"
	default:
		return fmt.Sprintf("ServiceState(%d)", state)
	}
}

func GetDecryptedFilesStatus(root string, serviceName string) ServiceDecryptionStatus {
	files, err := ResolveServiceFiles(root, serviceName, true)
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
	case decryptedFileCount >= 1 && decryptedFileCount == len(files):
		return All
	case decryptedFileCount >= 1 && decryptedFileCount < len(files):
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
