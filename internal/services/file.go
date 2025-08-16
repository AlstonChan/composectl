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

import "os"

type LineEnding int

const (
	Unknown LineEnding = iota
	LF
	CRLF
)

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
