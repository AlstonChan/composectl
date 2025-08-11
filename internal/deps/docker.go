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

package deps

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func parseMajorVersion(output string) (int, error) {
	// Extract first number group from output (e.g., "v0.26.1" → 0)
	fields := strings.Fields(output)
	for _, f := range fields {
		ver := strings.TrimPrefix(f, "v")
		if parts := strings.Split(ver, "."); len(parts) > 0 {
			if major, err := strconv.Atoi(parts[0]); err == nil {
				return major, nil
			}
		}
	}
	return 0, errors.New("no valid version number found")
}

func CheckDockerDeps(requiredBuildxMajor, requiredComposeMajor int) error {
	// 1. Docker Engine
	if _, err := CheckCommandExists("docker", "--version"); err != nil {
		return err
	}

	// 2. Docker Buildx
	buildxOut, err := CheckCommandExists("docker", "buildx", "version")
	if err != nil {
		return err
	}
	if major, err := parseMajorVersion(buildxOut); err != nil {
		return fmt.Errorf("failed to parse buildx version: %v", err)
	} else if major != requiredBuildxMajor {
		return fmt.Errorf("docker buildx major version %d required, found %d", requiredBuildxMajor, major)
	}

	// 3. Docker Compose
	composeOut, err := CheckCommandExists("docker", "compose", "version")
	if err != nil {
		return err
	}
	if major, err := parseMajorVersion(composeOut); err != nil {
		return fmt.Errorf("failed to parse compose version: %v", err)
	} else if major != requiredComposeMajor {
		return fmt.Errorf("docker compose major version %d required, found %d", requiredComposeMajor, major)
	}

	return nil
}
