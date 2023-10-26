// Copyright 2023 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	artifact_registry "go.linka.cloud/artifact-registry"
)

var (
	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version information and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", artifact_registry.Version)
			fmt.Printf("Commit: %s\n", artifact_registry.Commit)
			fmt.Printf("Date: %s\n", artifact_registry.Date)
			fmt.Printf("Repo: https://github.com/%s\n", artifact_registry.Repo)
		},
	}
)

func init() {
	rootCmd.AddCommand(cmdVersion)
}
