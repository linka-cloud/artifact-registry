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
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"go.linka.cloud/printer"

	"go.linka.cloud/artifact-registry/pkg/api"
	"go.linka.cloud/artifact-registry/pkg/slices"
)

var (
	repoGroup = &cobra.Group{ID: "1_repositories", Title: "Repository Commands:"}

	repoCmd = &cobra.Command{
		Use:               "repositories [registry]",
		Short:             "List repositories in the registry",
		Aliases:           []string{"repo", "repos"},
		Args:              cobra.ExactArgs(1),
		GroupID:           repoGroup.ID,
		PersistentPreRunE: setup,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := api.NewClient(registry, repository, opts...)
			if err != nil {
				return err
			}
			repos, err := c.Repositories(cmd.Context())
			if err != nil {
				return err
			}
			type Repo struct {
				Image         string     `json:"name" print:"IMAGE"`
				Type          string     `json:"type" print:"TYPE"`
				Size          int64      `json:"size" print:"SIZE"`
				LastUpdated   *time.Time `json:"lastUpdated" print:"LAST UPDATED"`
				Packages      int64      `json:"packages" print:"PACKAGES"`
				PackagesSize  int64      `json:"packagesSize" print:"PACKAGES SIZE"`
				MetadataFiles int64      `json:"metadataFiles" print:"METADATA FILES"`
				MetadataSize  int64      `json:"metadataSize" print:"METADATA SIZE"`
			}
			out := slices.Map(repos, func(v *api.Repository) Repo {
				return Repo{
					Image:         v.Name + ":" + v.Type,
					Type:          v.Type,
					Size:          v.Size,
					LastUpdated:   v.LastUpdated,
					MetadataFiles: v.Metadata.Count,
					MetadataSize:  v.Metadata.Size,
					Packages:      v.Packages.Count,
					PackagesSize:  v.Packages.Size,
				}
			})
			if err := printer.Print(
				out,
				printer.WithFormat(format),
				printer.WithYAMLMarshaler(yaml.Marshal),
				printer.WithFormatter("Size", formatSize),
				printer.WithFormatter("PackagesSize", formatSize),
				printer.WithFormatter("MetadataSize", formatSize),
			); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	repoCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (table, json, yaml)")
	repoCmd.RegisterFlagCompletionFunc("output", completeOutput)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddGroup(repoGroup)
}
