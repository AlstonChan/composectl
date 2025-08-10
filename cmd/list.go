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
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/docker"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
)

type Service struct {
	sequence int
	name     string
}
type ServiceOutput struct {
	Service
	dockerStatus string
	decrypted    bool
}

func processService(channel <-chan Service, result []ServiceOutput, counter *int32, repoRoot string) {
	for service := range channel {
		var serviceDirectory string = filepath.Join(repoRoot, config.DockerServicesDir, service.name)
		var serviceStatus string = "Unused"

		decrypted := services.IsEnvDecrypted(fmt.Sprintf("%s/.env", serviceDirectory))

		state, err := docker.GetServiceState(serviceDirectory)
		if err != nil {
			log.Fatal(err)
		}

		switch state {
		case docker.Running:
			serviceStatus = "Running"
		case docker.PartiallyRunning:
			serviceStatus = "Partial"
		case docker.Stopped:
			serviceStatus = "Stopped"
		}

		// Atomically get the next index
		idx := atomic.AddInt32(counter, 1) - 1
		result[idx] = ServiceOutput{dockerStatus: serviceStatus, decrypted: decrypted, Service: service}
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the selfhost repo with status",
	Run: func(cmd *cobra.Command, args []string) {
		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			log.Fatalf("Error resolving repo root: %v", err)
		}

		serviceList, err := services.ListAllService(repoRoot)
		if err != nil {
			log.Fatalf("Error listing services: %v", err)
		}

		fmt.Printf("Repo root: %s\n\n", repoRoot)
		if len(serviceList) == 0 {
			fmt.Println("No services found.")
			return
		}

		var serviceWg sync.WaitGroup
		var serviceChannel chan Service = make(chan Service)

		// The output data slice with pre-occupied capacity according to the
		// service list to avoid further dynamic allocating during service processing
		var serviceOutput []ServiceOutput = make([]ServiceOutput, len(serviceList))
		// Atomic counter to track the count of processed service in the serviceOutput
		var atomicCounter int32 = 0

		// Start worker goroutines
		var numWorkers int = runtime.NumCPU()
		for i := 1; i <= numWorkers; i++ {
			serviceWg.Add(1)
			go func() {
				defer serviceWg.Done()
				processService(serviceChannel, serviceOutput, &atomicCounter, repoRoot)
			}()
		}

		// Populate channel with services
		for index, service := range serviceList {
			serviceChannel <- Service{sequence: index + 1, name: service}
		}
		close(serviceChannel)
		serviceWg.Wait()

		sort.Slice(serviceOutput, func(i, j int) bool {
			return serviceOutput[i].sequence < serviceOutput[j].sequence
		})

		// Print final results
		for _, result := range serviceOutput {
			fmt.Printf("%2d - %-25s  Status: %-10s  Decrypted: %-5t\n",
				result.sequence, result.name, result.dockerStatus, result.decrypted)
		}
		fmt.Print("\n")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
