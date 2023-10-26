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
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.linka.cloud/grpc-toolkit/cli/clifmt"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/printer"
	"oras.land/oras-go/v2/registry/remote/credentials"

	artifact_registry "go.linka.cloud/artifact-registry"
	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
)

var (
	registry   string
	repository string
	user       string
	pass       string

	caFile   string
	insecure bool

	plainHTTP bool

	output string

	format printer.Format

	credsStore *credentials.DynamicStore
	caPool     *x509.CertPool

	opts []hclient.Option

	debug bool

	rootCmd = &cobra.Command{
		Use:          "lkar",
		Short:        "An OCI based Artifact Registry",
		SilenceUsage: true,
		Version:      artifact_registry.Version,
	}
)

func setup(cmd *cobra.Command, args []string) error {
	if debug {
		logger.SetDefault(logger.StandardLogger().SetLevel(logger.DebugLevel))
	}
	if caFile != "" {
		caPool = x509.NewCertPool()
		b, err := os.ReadFile(caFile)
		if err != nil {
			return fmt.Errorf("--ca-file: %w", err)
		}
		if !caPool.AppendCertsFromPEM(b) {
			return fmt.Errorf("--ca-file: no valid certificates found")
		}
		opts = append(opts, hclient.WithClientCA(caPool))
	}
	var err error
	if format, err = printer.ParseFormat(output); err != nil {
		return err
	}
	arg := args[0]
	parts := strings.SplitN(arg, "/", 2)
	registry = parts[0]
	if len(parts) > 1 {
		repository = parts[1]
	}
	credsStore, err = credentials.NewStoreFromDocker(credentials.StoreOptions{
		AllowPlaintextPut: true,
	})
	if err != nil {
		return err
	}
	if err := makeCreds(cmd.Context()); err != nil {
		return err
	}
	opts = append(opts, hclient.WithBasicAuth(user, pass), hclient.WithUserAgent(fmt.Sprintf("lkar/%s", artifact_registry.Version)))
	if plainHTTP {
		opts = append(opts, hclient.WithPlainHTTP())
	}
	if insecure {
		opts = append(opts, hclient.WithInsecure())
	}
	return nil
}

func makeCreds(ctx context.Context) error {
	if user != "" || pass != "" {
		return nil
	}
	// get credentials from the store
	creds, err := credsStore.Get(ctx, repoURL())
	if err != nil {
		return err
	}
	if repository == "" {
		user, pass = creds.Username, creds.Password
		return nil
	}
	if creds.Username == "" && creds.Password == "" {
		creds, err = credsStore.Get(ctx, registry)
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
	rootCmd.PersistentFlags().StringVarP(&caFile, "ca-file", "", "", "CA certificate file")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "Do not verify tls certificates")
	rootCmd.PersistentFlags().BoolVarP(&plainHTTP, "plain-http", "H", false, "Use http instead of https")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	logger.StandardLogger().Logger().SetFormatter(clifmt.New(clifmt.NoneTimeFormat))
	logger.SetDefault(logger.StandardLogger())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
