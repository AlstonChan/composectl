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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/AlstonChan/composectl/internal/config"
	"github.com/AlstonChan/composectl/internal/deps"
	"github.com/AlstonChan/composectl/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Service struct {
	sequence int
	name     string
}
type ServiceOutput struct {
	Service
	dockerStatus  string
	decryptStatus services.ServiceDecryptionStatus
}

func processService(channel <-chan Service, result []ServiceOutput, counter *int32, repoRoot string) {
	for service := range channel {
		var serviceDirectory string = filepath.Join(repoRoot, config.DockerServicesDir, service.name)

		decryptStatus := services.GetDecryptedFilesStatus(repoRoot, service.name)

		state, err := services.GetActiveServiceState(serviceDirectory)
		if err != nil {
			log.Fatal(err)
		}
		var serviceStatus string = services.GetServiceStatusString(state.ServiceState)
		if state.Label != "" {
			serviceStatus += " (" + state.Label + ")"
		}

		// Atomically get the next index
		idx := atomic.AddInt32(counter, 1) - 1
		result[idx] = ServiceOutput{dockerStatus: serviceStatus, decryptStatus: decryptStatus, Service: service}
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the self-host repo with status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := deps.CheckDockerDeps(0, 2); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if repoPath == "" {
			services.CreateLocalCacheDir(os.Getenv(config.ConfigDirEnv))
			if val := viper.GetString(CONFIG_REPO_PATH); val != "" {
				repoPath = val
			}
		}

		repoRoot, err := services.ResolveRepoRoot(repoPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving repo root: %v\n", err)
		}

		serviceList, err := services.ListAllService(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing services: %v\n", err)
		}

		if len(serviceList) == 0 {
			fmt.Fprintln(os.Stderr, "No services found.")
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
			fmt.Printf("%2d - %-23s  Status: %-16s  Decrypted: %-6s\n",
				result.sequence, result.name, result.dockerStatus, services.GetDecryptedStatusString(result.decryptStatus))
		}
		fmt.Print("\n")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
