// Copyright 2024 Linka Cloud  All rights reserved.
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
	"github.com/spf13/cobra"
	"os"
	"text/template"
)

var (
	domain     string
	deployMode string
	repoMode   string
)

func main() {
	cmd := cobra.Command{
		Use:  "build-docs <template-path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dm, err := ParseDeployMode(deployMode)
			if err != nil {
				return err
			}
			rm, err := ParseRepoMode(repoMode)
			if err != nil {
				return err
			}
			b, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			tpl := template.Must(template.New("docs").Parse(string(b)))
			return tpl.Execute(os.Stdout, &TemplateVariables{Domain: domain, DeployMode: dm, RepoMode: rm})
		},
	}
	cmd.Flags().StringVarP(&domain, "domain", "d", "example.org", "domain name")
	cmd.Flags().StringVarP(&deployMode, "deploy-mode", "m", "", "deploy mode")
	cmd.Flags().StringVarP(&repoMode, "repo-mode", "r", "", "repo mode")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
