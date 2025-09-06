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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const JSON_METADATA_VERSION = "1.0"

type Metadata struct {
	Version     string   `json:"version"`
	Service     string   `json:"service"`
	Timestamp   string   `json:"timestamp"`
	ComposeFile string   `json:"compose_file"`
	Volumes     []Volume `json:"volumes"`
}
type Volume struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

var genBackupMetaCmd = &cobra.Command{
	Use:   "gen-backup-meta",
	Short: "Generate the json metadata file for a backup tarball",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		input, _ := cmd.Flags().GetString("input")
		serviceName, _ := cmd.Flags().GetString("service")

		var fullInputPath, err = filepath.Abs(input)
		if err != nil {
			return fmt.Errorf("unable to parse input docker compose file path: %v", err)
		}

		fullOutputPath, err := filepath.Abs(output)
		if err != nil {
			return fmt.Errorf("unable to parse output tarball file path: %v", err)
		}

		// Get the docker compose file from input string
		var dockerComposeFilePath string = fullInputPath
		var inputPathExtension = filepath.Ext(fullInputPath)
		if inputPathExtension != ".yml" && inputPathExtension != ".yaml" {
			if inputPathExtension != "" {
				return fmt.Errorf("is the input file a docker compose yaml file? (%s)", filepath.Base(dockerComposeFilePath))
			}

			matchingComposeFile, err := services.FindComposeFiles(fullInputPath)
			if err != nil {
				return fmt.Errorf("unable to locate docker compose file: %v", err)
			}

			dockerComposeFilePath = filepath.Join(fullInputPath, matchingComposeFile[0])
		}

		// Read the docker compose file content
		data, err := os.ReadFile(dockerComposeFilePath)
		if err != nil {
			return fmt.Errorf("unable to read the docker compose file: %v", err)
		}

		// Parsing the docker compose file by unmarshaling it into generic map
		// because we don't know how this docker compose files looks like for sure
		var compose map[string]any
		if err := yaml.Unmarshal(data, &compose); err != nil {
			return fmt.Errorf("unable to parse the docker compose file: %v", err)
		}

		services, ok := compose["services"].(map[string]any)
		if !ok {
			return fmt.Errorf("the docker compose file does not have a service section")
		}

		// Determine which volume contains the backup path
		var tempVolumeMappings []Volume = make([]Volume, 0)

		// Get the specified service first that defaulted to "backup"
		if service, ok := services[serviceName]; ok {
			fmt.Printf("Service \"%s\" found\n", serviceName)

			serviceMap, ok := service.(map[string]any)
			if !ok {
				return fmt.Errorf("unable to parse the service \"%s\" correctly", serviceName)
			}

			// Get the volumes section of the service
			if volumes, ok := serviceMap["volumes"]; ok {
				if volume, ok := volumes.([]any); ok {
					fmt.Printf("%d Volumes found\n", len(volume))

					// For every volume entry
					for i, v := range volume {
						if volStr, ok := v.(string); ok {
							var volumePath []string = strings.Split(volStr, ":")

							// Check if the volume mount is the one we want. so if it was mounted
							// to /archive, it is for local backup which we will ignore, and we also
							// ignore mounts to docker socket
							if strings.HasPrefix(volumePath[1], "/archive") || strings.HasPrefix(volumePath[1], "/var/run/docker.sock") {
								continue
							}

							// Temporarily save the data to a slice
							tempVolumeMappings = append(tempVolumeMappings, Volume{Name: volumePath[0], Path: volumePath[1]})
						} else {
							fmt.Printf("  %d: (non-string value: %#v)\n", i+1, v)
						}
					}

				} else {
					return fmt.Errorf("volumes field exists but is not a list")
				}
			} else {
				return fmt.Errorf("volumes \"%s\" not found", volumes)
			}
		} else {
			return fmt.Errorf("service \"%s\" not found", serviceName)
		}

		// Parse the volumes section to get the volume name and set the volumes property
		volumes, ok := compose["volumes"].(map[string]any)
		if !ok {
			return fmt.Errorf("the docker compose file does not have a volumes section")
		}

		var volumeMappings []Volume = make([]Volume, 0)

		// Loop through each entry in volumes. We will filter out local volume mount here
		// and also determine the actual volume name and the path of the data

		// This volume does not have a name defined, determine the name using the directory name
		// This could be used later, so we defined it here
		var serviceDirName string = filepath.Base(filepath.Dir(dockerComposeFilePath))

		for _, v := range tempVolumeMappings {
			// Check if the volume is defined as a named volume
			if volume, ok := volumes[v.Name]; ok {
				if volumeMap, ok := volume.(map[string]any); ok {
					// Get the volume name
					if volumes, ok := volumeMap["name"]; ok {
						if volumeName, ok := volumes.(string); ok {
							volumeMappings = append(volumeMappings, Volume{Name: volumeName, Path: v.Path})
						} else {
							fmt.Println("volumes field cannot be parsed correctly")
						}
					} else {
						// Append the directory name of the docker compose file and concat them with a underscore
						volumeMappings = append(volumeMappings, Volume{Name: serviceDirName + "_" + v.Name, Path: v.Path})
					}
				} else {
					fmt.Printf("volume %q don't have any settings\n", v.Name)

					// Append the directory name of the docker compose file and concat them with a underscore
					volumeMappings = append(volumeMappings, Volume{Name: serviceDirName + "_" + v.Name, Path: v.Path})
				}
			} else {
				// Happens because the volume is local mount
				fmt.Printf("volumes \"%s\" not found in volumes section\n", v)
			}
		}

		// we have to determine the service name. We will try to get it from the docker compose
		// name entry first before defaulting to serviceDirName
		dockerServiceName, ok := compose["name"].(string)

		var service string
		if ok {
			service = dockerServiceName
		} else {
			service = serviceDirName
		}

		// Create the metadata
		var metadata Metadata = Metadata{
			Version:     JSON_METADATA_VERSION,
			Service:     service,
			Timestamp:   time.Now().Format(time.RFC3339),
			ComposeFile: filepath.Base(dockerComposeFilePath),
			Volumes:     volumeMappings,
		}

		var targetFile string = fullOutputPath
		// Append backup.json if it does not already include a json file in the output string
		if filepath.Ext(fullOutputPath) != ".json" {
			targetFile = filepath.Join(fullOutputPath, "backup.json")
		}

		// Open the file, create it if it does not exists, and overwrite everything
		file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("unable to open/create to file %v", targetFile)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ") // pretty print
		if err := encoder.Encode(metadata); err != nil {
			return fmt.Errorf("unable to write to file %v", targetFile)
		}

		fmt.Printf("Backup metadata written to %s for service %s\n", fullOutputPath, service)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genBackupMetaCmd)
	genBackupMetaCmd.Flags().StringP("output", "o", "", "The directory path to write the metadata to")
	genBackupMetaCmd.Flags().StringP("input", "i", "", "The path of the docker compose file to generate the metadata for")
	genBackupMetaCmd.Flags().String("service", "backup", "The docker service that does the backup")
	genBackupMetaCmd.MarkFlagRequired("output")
	genBackupMetaCmd.MarkFlagRequired("input")
}
