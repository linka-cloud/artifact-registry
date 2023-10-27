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
	"os"

	"github.com/spf13/cobra"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/packages/apk"
	"go.linka.cloud/artifact-registry/pkg/packages/deb"
	"go.linka.cloud/artifact-registry/pkg/packages/helm"
	"go.linka.cloud/artifact-registry/pkg/packages/rpm"
)

func newPkgPushCmd(typ string) *cobra.Command {
	use := fmt.Sprintf("push [repository] [path]")
	index := 1
	var client func(args []string) (packages.Pusher, error)
	switch typ {
	case apk.Name:
		use = fmt.Sprintf("push [repository] [branch] [apk-repository] [path]")
		index = 3
		client = func(args []string) (packages.Pusher, error) {
			return apk.NewClient(registry, repository, args[1], args[2], opts...)
		}
	case deb.Name:
		use = fmt.Sprintf("push [repository] [distribution] [component] [path]")
		index = 3
		client = func(args []string) (packages.Pusher, error) {
			return deb.NewClient(registry, repository, args[1], args[2], opts...)
		}
	case rpm.Name:
		client = func(args []string) (packages.Pusher, error) {
			return rpm.NewClient(registry, repository, opts...)
		}
	case helm.Name:
		client = func(args []string) (packages.Pusher, error) {
			return helm.NewClient(registry, repository, opts...)
		}
	default:
		panic(fmt.Sprintf("unknown package type %s", typ))
	}
	return &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf("Push %s package to the repository", typ),
		Aliases: []string{"put", "create", "upload"},
		Args:    cobra.ExactArgs(index + 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			f, err := os.Open(args[index])
			if err != nil {
				return err
			}
			defer f.Close()
			i, err := f.Stat()
			if err != nil {
				return err
			}
			c, err := client(args)
			if err != nil {
				return err
			}
			pw := newProgressReader(f, i.Size())
			go pw.Run(ctx)
			if err := c.Push(ctx, pw); err != nil {
				return err
			}
			return nil
		},
	}
}
