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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.linka.cloud/printer"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

var (
	registry   string
	repository string
	user       string
	pass       string

	insecure bool

	plainHTTP bool

	output string

	format printer.Format

	credsStore *credentials.DynamicStore

	rootCmd = &cobra.Command{
		Use:               "lkar",
		SilenceUsage:      true,
		PersistentPreRunE: setup,
	}
)

func setup(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	arg := args[0]
	parts := strings.Split(arg, "/")
	registry = parts[0]
	if len(parts) > 1 {
		repository = strings.Join(parts[1:], "/")
	}
	var err error
	if format, err = printer.ParseFormat(output); err != nil {
		return err
	}
	credsStore, err = credentials.NewStoreFromDocker(credentials.StoreOptions{
		AllowPlaintextPut: true,
	})
	if err != nil {
		return err
	}
	if user != "" || pass != "" {
		return nil
	}
	// get credentials from the store
	creds, err := credsStore.Get(cmd.Context(), repoURL())
	if err != nil {
		return err
	}
	if repository == "" {
		return nil
	}
	if creds.Username == "" && creds.Password == "" {
		creds, err = credsStore.Get(cmd.Context(), registry)
		if err != nil {
			return err
		}
	}
	user, pass = creds.Username, creds.Password
	return nil
}

func completeOutput(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch {
	case strings.HasPrefix("table", toComplete):
		return []string{"table"}, cobra.ShellCompDirectiveNoFileComp
	case strings.HasPrefix("json", toComplete):
		return []string{"json"}, cobra.ShellCompDirectiveNoFileComp
	case strings.HasPrefix("yaml", toComplete):
		return []string{"yaml"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Username")
	rootCmd.PersistentFlags().StringVarP(&pass, "pass", "p", "", "Password")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "Do not verify tls certificates")
	rootCmd.PersistentFlags().BoolVarP(&plainHTTP, "plain-http", "H", false, "Use http instead of https")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
