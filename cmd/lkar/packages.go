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

	"go.linka.cloud/artifact-registry/pkg/packages/apk"
	"go.linka.cloud/artifact-registry/pkg/packages/deb"
	"go.linka.cloud/artifact-registry/pkg/packages/helm"
	"go.linka.cloud/artifact-registry/pkg/packages/rpm"
)

var PkgGroup = &cobra.Group{ID: "2_packages", Title: "Package Commands:"}

func newPkgCmd(typ string) *cobra.Command {
	pkgCmd := &cobra.Command{
		Use:               typ,
		Short:             fmt.Sprintf("Manage %s packages", typ),
		GroupID:           PkgGroup.ID,
		PersistentPreRunE: setup,
	}
	pkgCmd.AddCommand(
		newPkgListCmd(typ),
		newPkgPushCmd(typ),
		newPkgPullCmd(typ),
		newPkgDeleteCmd(typ),
		newPkgSetupCmd(typ),
	)
	return pkgCmd
}

func init() {
	rootCmd.AddGroup(PkgGroup)
	for _, v := range []string{apk.Name, deb.Name, rpm.Name, helm.Name} {
		rootCmd.AddCommand(newPkgCmd(v))
	}
}
