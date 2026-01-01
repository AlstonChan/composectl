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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/client"
)

func parseMajorVersion(output string) (int, error) {
	// Extract first number group from output (e.g., "v0.26.1" → 0)
	fields := strings.FieldsSeq(output)
	for f := range fields {
		ver := strings.TrimPrefix(f, "v")
		if parts := strings.Split(ver, "."); len(parts) > 0 {
			if major, err := strconv.Atoi(parts[0]); err == nil {
				return major, nil
			}
		}
	}
	return 0, fmt.Errorf("no valid version number found")
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

// GetDockerClient runs the Docker dependency checks and returns an initialized
// Docker client. This centralizes the Docker connectivity checks and creation
// so callers can reuse the same logic without duplicating client creation.
//
// The caller is responsible for closing the returned client when it is no longer
// needed to avoid leaking resources, for example:
//
//	dockerClient, err := deps.GetDockerClient(requiredBuildxMajor, requiredComposeMajor)
//	if err != nil {
//	    return err
//	}
//	defer dockerClient.Close()
func GetDockerClient(requiredBuildxMajor, requiredComposeMajor int) (*client.Client, error) {
	if err := CheckDockerDeps(requiredBuildxMajor, requiredComposeMajor); err != nil {
		return nil, err
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// Verify daemon is reachable early so callers get a helpful error message
	// instead of failing later during an operation.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := dockerClient.Ping(ctx); err != nil {
		_ = dockerClient.Close()
		return nil, fmt.Errorf("unable to connect to Docker daemon: %v", err)
	}

	return dockerClient, nil
}
