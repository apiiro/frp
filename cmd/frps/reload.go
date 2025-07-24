// Copyright 2021 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reloadCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Trigger authentication config reload for running frps process",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find frps process
		out, err := exec.Command("pgrep", "-f", "frps").Output()
		if err != nil {
			fmt.Println("Error: No running frps process found")
			return nil
		}

		pids := strings.Fields(strings.TrimSpace(string(out)))
		if len(pids) == 0 {
			fmt.Println("Error: No running frps process found")
			return nil
		}

		if len(pids) > 1 {
			fmt.Printf("Warning: Multiple frps processes found (%s). Sending signal to all.\n", strings.Join(pids, ", "))
		}

		for _, pidStr := range pids {
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				fmt.Printf("Error: Invalid PID %s\n", pidStr)
				continue
			}

			process, err := os.FindProcess(pid)
			if err != nil {
				fmt.Printf("Error: Could not find process %d: %v\n", pid, err)
				continue
			}

			err = process.Signal(syscall.SIGUSR1)
			if err != nil {
				fmt.Printf("Error: Could not send signal to process %d: %v\n", pid, err)
				continue
			}

			fmt.Printf("Successfully sent reload signal to frps process %d\n", pid)
		}

		fmt.Println("Check frps logs to verify reload status")
		return nil
	},
}
