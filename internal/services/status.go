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
	"os/exec"
	"strings"
)

func IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	for name := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if strings.Contains(name, serviceName) {
			return true
		}
	}
	return false
}

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
