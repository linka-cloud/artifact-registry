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
	"sort"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"go.linka.cloud/printer"

	"go.linka.cloud/artifact-registry/pkg/api"
	"go.linka.cloud/artifact-registry/pkg/slices"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

func newPkgListCmd(typ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("list [repository]"),
		Short:   fmt.Sprintf("List %s packages in the repository", typ),
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := api.NewClient(registry, repository, opts...)
			if err != nil {
				return err
			}
			pkgs, err := c.Packages(ctx, typ)
			if err != nil {
				return err
			}

			type Package struct {
				Name    string `json:"name" print:"NAME"`
				Version string `json:"version" print:"VERSION"`
				Arch    string `json:"arch" print:"ARCH"`
				Size    int64  `json:"size" print:"SIZE"`
				Path    string `json:"path" print:"PATH"`
			}
			out := slices.Map(pkgs, func(v storage.Artifact) Package {
				return Package{
					Name:    v.Name(),
					Version: v.Version(),
					Arch:    v.Arch(),
					Size:    v.Size(),
					Path:    v.Path(),
				}
			})
			sort.Slice(out, func(i, j int) bool {
				return sort.StringsAreSorted([]string{out[i].Name, out[j].Name})
			})
			if err := printer.Print(
				out,
				printer.WithFormat(format),
				printer.WithFormatter("Size", formatSize),
				printer.WithYAMLMarshaler(yaml.Marshal),
			); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.RegisterFlagCompletionFunc("output", completeOutput)
	return cmd
}
