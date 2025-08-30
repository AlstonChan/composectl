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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

type LineEnding int

const (
	Unknown LineEnding = iota
	LF
	CRLF
)

// Regex for matching file names
// Matches:
// - docker-compose.yml
// - docker_compose.<anything>.yml
// - compose.yml
// - compose.<anything>.yml
var dockerComposeFileRegex = regexp.MustCompile(`^(docker[-_]compose|compose)(?:\.(.+))?\.yml$`)

func DetectLineEnding(path string) (LineEnding, error) {
	file, err := os.Open(path)
	if err != nil {
		return Unknown, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return Unknown, err
	}
	size := info.Size()
	if size == 0 {
		return Unknown, nil // empty file
	}

	readSize := int64(2)
	if size < 2 {
		readSize = size
	}

	buf := make([]byte, readSize)
	_, err = file.ReadAt(buf, size-readSize)
	if err != nil {
		return Unknown, err
	}

	if len(buf) >= 2 && buf[len(buf)-2] == '\r' && buf[len(buf)-1] == '\n' {
		return CRLF, nil
	} else if buf[len(buf)-1] == '\n' {
		return LF, nil
	}

	return Unknown, nil
}

// Recursively finds Docker Compose files under startingPath
func FindComposeFiles(startingPath string) ([]string, error) {
	var results []string

	err := filepath.WalkDir(startingPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If we cannot access a directory, just skip it
			return nil
		}
		if !d.IsDir() && dockerComposeFileRegex.MatchString(d.Name()) {
			relPath, err := filepath.Rel(startingPath, path)
			if err != nil {
				return err
			}
			results = append(results, relPath)
		}
		return nil
	})

	return results, err
}

// ExtractComposeVariant extracts the "middle part" of a docker compose filename.
// Returns "" if it's the base file (compose.yml or docker-compose.yml).
// Returns error if it doesn't match known patterns.
func ExtractComposeVariant(path string) (string, error) {
	// Just take the file name, not the full path
	filename := filepath.Base(path)

	m := dockerComposeFileRegex.FindStringSubmatch(filename)
	if m == nil {
		return "", errors.New("not a recognized compose file")
	}

	// m[2] contains the optional "middle part"
	return m[2], nil
}
