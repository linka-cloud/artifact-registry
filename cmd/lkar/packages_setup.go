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
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/packages/apk"
	"go.linka.cloud/artifact-registry/pkg/packages/deb"
	"go.linka.cloud/artifact-registry/pkg/packages/helm"
	"go.linka.cloud/artifact-registry/pkg/packages/rpm"
)

func newPkgSetupCmd(typ string) *cobra.Command {
	var (
		force  bool
		use    string
		args   int
		client func(ctx context.Context, scheme, name string, args []string) (packages.Setuper, error)
	)
	switch typ {
	case apk.Name:
		use = fmt.Sprintf("setup [repository] [branch] [apk-repository]")
		args = 3
		client = func(ctx context.Context, scheme, name string, args []string) (packages.Setuper, error) {
			return apk.NewClient(registry, repository, args[1], args[2], opts...)
		}
	case deb.Name:
		use = fmt.Sprintf("setup [repository] [distribution] [component]")
		args = 3
		client = func(ctx context.Context, scheme, name string, args []string) (packages.Setuper, error) {
			return deb.NewClient(registry, repository, args[1], args[2], opts...)
		}
	case rpm.Name:
		use = fmt.Sprintf("setup [repository]")
		args = 1
		client = func(ctx context.Context, scheme, name string, args []string) (packages.Setuper, error) {
			return rpm.NewClient(registry, repository, opts...)
		}
	case helm.Name:
		use = fmt.Sprintf("setup [repository]")
		args = 1
		client = func(ctx context.Context, scheme, name string, args []string) (packages.Setuper, error) {
			return helm.NewClient(registry, repository, opts...)
		}
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Setup %s repository on the machine", typ),
		Args:  cobra.ExactArgs(args),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if runtime.GOOS != "linux" && typ != "helm" {
				return fmt.Errorf("command only supported on Linux")
			}
			// Check if the user has root privileges
			if os.Geteuid() != 0 && typ != "helm" {
				return fmt.Errorf("please run as root or sudo")
			}
			scheme := "https"
			if plainHTTP {
				scheme = "http"
			}
			name := strings.Replace(repository, "/", "-", -1)
			if repository == "" {
				name = strings.Replace(strings.Split(registry, ":")[0], ".", "-", -1)
			}
			c, err := client(ctx, scheme, name, args)
			if err != nil {
				return err
			}
			return c.SetupLocal(ctx, force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force setup")
	return cmd
}
