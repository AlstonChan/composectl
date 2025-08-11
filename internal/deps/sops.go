package deps

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func CheckSops() error {
	cmd := exec.Command("sops", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return errors.New("sops not found in PATH or not executable")
	}

	// Example output: "sops 3.10.2"
	fields := strings.Fields(out.String())
	if len(fields) < 2 {
		return fmt.Errorf("unexpected sops version output: %q", out.String())
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
