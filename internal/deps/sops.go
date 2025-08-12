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
	"fmt"
	"strconv"
	"strings"
)

func CheckSops() error {
	sops, err := CheckCommandExists("sops", "--version")
	if err != nil {
		return err
	}

	// Example output: "sops 3.10.2"
	fields := strings.Fields(sops)
	if len(fields) < 2 {
		return fmt.Errorf("unexpected sops version output: %q", sops)
	}

	version := fields[1] // "3.10.2"
	majorStr := strings.SplitN(version, ".", 2)[0]
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return fmt.Errorf("failed to parse sops version: %q", version)
	}

	if major != 3 {
		return fmt.Errorf("sops major version must be 3, found %d", major)
	}

	return nil
}
