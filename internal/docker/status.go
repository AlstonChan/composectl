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

package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type ServiceState int

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

func GetServiceState(projectDir string) (ServiceState, error) {
	cmd := exec.Command("docker", "compose", "ps", "--format", "json")
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
