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
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.linka.cloud/env"
	"go.linka.cloud/grpc-toolkit/logger"

	artifact_registry "go.linka.cloud/artifact-registry"
	"go.linka.cloud/artifact-registry/pkg/registry"
	"go.linka.cloud/artifact-registry/pkg/server"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const (
	EnvAddr         = "ARTIFACT_REGISTRY_ADDRESS"
	EnvBackend      = "ARTIFACT_REGISTRY_BACKEND"
	EnvKey          = "ARTIFACT_REGISTRY_AES_KEY"
	EnvDomain       = "ARTIFACT_REGISTRY_DOMAIN"
	EnvNoHTTPS      = "ARTIFACT_REGISTRY_NO_HTTPS"
	EnvInsecure     = "ARTIFACT_REGISTRY_INSECURE"
	EnvTagArtifacts = "ARTIFACT_REGISTRY_TAG_ARTIFACTS"
	EnvClientCA     = "ARTIFACT_REGISTRY_CLIENT_CA"
	EnvTLSCert      = "ARTIFACT_REGISTRY_TLS_CERT"
	EnvTLSKey       = "ARTIFACT_REGISTRY_TLS_KEY"
	EnvDisableUI    = "ARTIFACT_REGISTRY_DISABLE_UI"

	EnvProxy         = "ARTIFACT_REGISTRY_PROXY"
	EnvProxyNoHTTPS  = "ARTIFACT_REGISTRY_PROXY_NO_HTTPS"
	EnvProxyInsecure = "ARTIFACT_REGISTRY_PROXY_INSECURE"
	EnvProxyClientCA = "ARTIFACT_REGISTRY_PROXY_CLIENT_CA"
	EnvProxyUser     = "ARTIFACT_REGISTRY_PROXY_USER"
	EnvProxyPassword = "ARTIFACT_REGISTRY_PROXY_PASSWORD"
)

var (
	addr = ":9887"

	backend = "docker.io"

	domain = ""

	aesKey = ""

	noHTTPS = false

	insecure = false

	tagPerArtifact = false

	key, cert string

	clientCA string

	disableUI = false

	proxyAddr     string
	proxyNoHTTPS  = false
	proxyInsecure = false
	proxyClientCA string
	proxyUser     string
	proxyPassword string

	debug bool

	cmd = &cobra.Command{
		Use:          "lkard (repository)",
		Short:        "An OCI based Artifact Registry",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		Version:      artifact_registry.Version,
		Run: func(cmd *cobra.Command, args []string) {
			if aesKey == "" {
				logrus.Fatalf("environment variable $%s must be set", EnvKey)
			}
			var repo string
			if len(args) > 0 {
				repo = args[0]
			}
			// TODO(adphi): validate host
			ropts := []registry.Option{
				registry.WithProxy(proxyAddr),
				registry.WithProxyUser(proxyUser),
				registry.WithProxyPassword(proxyPassword),
			}
			if debug {
				ropts = append(ropts, registry.WithDebug())
			}
			if noHTTPS {
				ropts = append(ropts, registry.WithPlainHTTP())
			}
			if insecure {
				ropts = append(ropts, registry.WithInsecure())
			}
			if clientCA != "" {
				p := x509.NewCertPool()
				b, err := os.ReadFile(clientCA)
				if err != nil {
					logger.C(cmd.Context()).Fatal(err)
				}
				if !p.AppendCertsFromPEM(b) {
					logger.C(cmd.Context()).Fatal(err)
				}
				ropts = append(ropts, registry.WithClientCA(p))
			}
			if proxyNoHTTPS {
				ropts = append(ropts, registry.WithProxyPlainHTTP())
			}
			if proxyInsecure {
				ropts = append(ropts, registry.WithProxyInsecure())
			}
			if proxyClientCA != "" {
				p := x509.NewCertPool()
				b, err := os.ReadFile(proxyClientCA)
				if err != nil {
					logger.C(cmd.Context()).Fatal(err)
				}
				if !p.AppendCertsFromPEM(b) {
					logger.C(cmd.Context()).Fatal(err)
				}
				ropts = append(ropts, registry.WithProxyClientCA(p))
			}
			opts := []storage.Option{
				storage.WithHost(backend),
				storage.WithRepo(repo),
				storage.WithRegistryOptions(ropts...),
			}
			if tagPerArtifact {
				opts = append(opts, storage.WithArtifactTags())
			}
			if strings.HasSuffix(backend, "docker.io") && proxyAddr == "" {
				logger.C(cmd.Context()).Warnf("using docker.io as backend without proxy is not recommended")
				logger.C(cmd.Context()).Warnf("the rate limit of 100 requests per 6 hours is very easy to reach using this tool")
			}
			if err := server.Run(cmd.Context(), addr, aesKey, backend, domain, repo, cert, key, disableUI, opts...); err != nil {
				logger.C(cmd.Context()).Fatal(err)
			}
		},
	}
	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version informations and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", artifact_registry.Version)
			fmt.Printf("Commit: %s\n", artifact_registry.Commit)
			fmt.Printf("Date: %s\n", artifact_registry.Date)
			fmt.Printf("Repo: https://github.com/%s\n", artifact_registry.Repo)
		},
	}
)

func main() {
	cmd.AddCommand(cmdVersion)
	cmd.Flags().StringVar(&addr, "addr", env.GetDefault(EnvAddr, addr), "address to listen on [$"+EnvAddr+"]")
	cmd.Flags().StringVar(&backend, "backend", env.GetDefault(EnvBackend, backend), "registry backend hostname (and port if not 443 or 80) [$"+EnvBackend+"]")
	cmd.Flags().StringVar(&aesKey, "aes-key", env.GetDefault(EnvKey, aesKey), "AES key to encrypt the repositories keys [$"+EnvKey+"]")
	cmd.Flags().StringVar(&domain, "domain", env.GetDefault(EnvDomain, domain), "domain to use to serve the repositories as subdomains [$"+EnvDomain+"]")
	cmd.Flags().BoolVar(&noHTTPS, "no-https", env.GetDefault(EnvNoHTTPS, noHTTPS), "disable backend registry client https [$"+EnvNoHTTPS+"]")
	cmd.Flags().BoolVar(&insecure, "insecure", env.GetDefault(EnvInsecure, insecure), "disable backend registry client tls verification [$"+EnvInsecure+"]")
	cmd.Flags().BoolVar(&tagPerArtifact, "tag-artifacts", env.GetDefault(EnvTagArtifacts, tagPerArtifact), "tag artifacts manifests [$"+EnvTagArtifacts+"]")
	cmd.Flags().StringVar(&clientCA, "client-ca", env.Get[string](EnvClientCA), "tls client certificate authority [$"+EnvClientCA+"]")
	cmd.Flags().StringVar(&cert, "tls-cert", env.Get[string](EnvTLSCert), "tls certificate [$"+EnvTLSCert+"]")
	cmd.Flags().StringVar(&key, "tls-key", env.Get[string](EnvTLSKey), "tls key [$"+EnvTLSKey+"]")
	cmd.Flags().BoolVar(&disableUI, "disable-ui", env.GetDefault(EnvDisableUI, disableUI), "disable the Web UI [$"+EnvDisableUI+"]")

	cmd.Flags().StringVar(&proxyAddr, "proxy", env.GetDefault(EnvProxy, proxyAddr), "proxy backend registry hostname (and port if not 443 or 80) [$"+EnvProxy+"]")
	cmd.Flags().BoolVar(&proxyNoHTTPS, "proxy-no-https", env.GetDefault(EnvProxyNoHTTPS, noHTTPS), "disable proxy registry client https [$"+EnvProxyNoHTTPS+"]")
	cmd.Flags().BoolVar(&proxyInsecure, "proxy-insecure", env.GetDefault(EnvProxyInsecure, insecure), "disable proxy registry client tls verification [$"+EnvProxyInsecure+"]")
	cmd.Flags().StringVar(&proxyClientCA, "proxy-client-ca", env.Get[string](EnvProxyClientCA), "proxy tls client certificate authority [$"+EnvProxyClientCA+"]")
	cmd.Flags().StringVar(&proxyUser, "proxy-user", env.GetDefault(EnvProxyUser, proxyUser), "proxy registry user [$"+EnvProxyUser+"]")
	cmd.Flags().StringVar(&proxyPassword, "proxy-password", env.GetDefault(EnvProxyPassword, proxyPassword), "proxy registry password [$"+EnvProxyPassword+"]")

	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")

	if debug {
		logger.SetDefault(logger.StandardLogger().SetLevel(logger.DebugLevel))
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
