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

	"go.linka.cloud/artifact-registry/pkg/packages/apk"
	"go.linka.cloud/artifact-registry/pkg/packages/deb"
	"go.linka.cloud/artifact-registry/pkg/packages/rpm"
)

func newPkgSetupCmd(typ string) *cobra.Command {
	var (
		force bool
		use   string
		args  int
		setup func(ctx context.Context, scheme string, args []string) error
	)
	var prefix string
	switch strings.Split(registry, ".")[0] {
	case "apk", "deb", "rpm":
		prefix = "/"
	default:
		prefix = "/" + typ + "/"
	}
	switch typ {
	case "apk":
		use = fmt.Sprintf("setup [repository] [branch] [apk-repository]")
		args = 3
		setup = func(ctx context.Context, scheme string, args []string) error {
			return apk.Setup(ctx, apk.SetupArgs{
				User:       user,
				Password:   pass,
				Scheme:     scheme,
				Host:       registry,
				Path:       prefix + repository,
				Branch:     args[1],
				Repository: args[2],
			}, force)
		}
	case "deb":
		use = fmt.Sprintf("setup [repository] [distribution] [component]")
		args = 3
		setup = func(ctx context.Context, scheme string, args []string) error {
			return deb.Setup(ctx, deb.SetupArgs{
				User:      user,
				Password:  pass,
				Scheme:    scheme,
				Host:      registry,
				Path:      prefix + repository,
				Name:      strings.Replace(repository, "/", "-", -1),
				Dist:      args[1],
				Component: args[2],
			}, force)
		}
	case "rpm":
		use = fmt.Sprintf("setup [repository]")
		args = 1
		setup = func(ctx context.Context, scheme string, args []string) error {
			return rpm.Setup(ctx, rpm.SetupArgs{
				User:     user,
				Password: pass,
				Scheme:   scheme,
				Host:     registry,
				Path:     prefix + repository,
			}, force)
		}
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Setup %s repository on the machine", typ),
		Args:  cobra.ExactArgs(args),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if runtime.GOOS != "linux" {
				return fmt.Errorf("command only supported on Linux")
			}
			// Check if the user has root privileges
			if os.Geteuid() != 0 {
				return fmt.Errorf("please run as root or sudo")
			}
			scheme := "https"
			if plainHTTP {
				scheme = "http"
			}
			return setup(ctx, scheme, args)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force setup")
	return cmd
}
