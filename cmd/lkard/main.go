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
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"go.linka.cloud/env"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/react"

	artifact_registry "go.linka.cloud/artifact-registry"
	"go.linka.cloud/artifact-registry/pkg/api"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/registry"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/ui"
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

	cmd = &cobra.Command{
		Use:          "artifact-registry (repository)",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
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
			if err := run(cmd.Context(), repo, opts...); err != nil {
				logger.C(cmd.Context()).Fatal(err)
			}
		},
	}
	cmdVersion = &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s (%s)\n", artifact_registry.Version, artifact_registry.BuildDate)
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

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func wrap(w http.ResponseWriter) *wrapWriter {
	var buf bytes.Buffer
	return &wrapWriter{ResponseWriter: w, body: &buf, w: io.MultiWriter(w, &buf)}
}

type wrapWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   *bytes.Buffer
	w      io.Writer
}

func (w *wrapWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		return 0, err
	}
	w.size += n
	return n, nil
}

func run(ctx context.Context, repo string, opts ...storage.Option) error {
	if aesKey == "" {
		return fmt.Errorf("environment variable $%s must be set", EnvKey)
	}
	logger.C(ctx).Infof("initializing artifact registry using backend %s", backend)
	router := mux.NewRouter().StrictSlash(true)
	k := sha256.Sum256([]byte(aesKey))
	ctx = storage.WithOptions(ctx, append(opts, storage.WithKey(k[:]))...)

	if !disableUI {
		uih, err := react.NewHandler(ui.UI, "build")
		if err != nil {
			return err
		}
		router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", uih))
		router.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui", http.StatusFound)
		})
	}

	router.Path("/_/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := api.Init(ctx, router, domain, repo); err != nil {
		return err
	}
	if err := packages.Init(ctx, router, domain, repo); err != nil {
		return err
	}

	if err := router.Walk(func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error { return r.GetError() }); err != nil {
		return err
	}

	s := http.Server{
		BaseContext: func(lis net.Listener) context.Context {
			return ctx
		},
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrap := wrap(w)
			start := time.Now()
			remote := strings.Split(r.RemoteAddr, ":")[0]
			for _, v := range r.Header["X-Forwarded-For"] {
				if ip := net.ParseIP(v); ip != nil && !ip.IsPrivate() {
					remote = v
					break
				}
			}
			log := logger.C(r.Context()).WithFields(
				"method", r.Method,
				"path", r.URL.Path,
				"remote", remote,
				"userAgent", r.UserAgent(),
			)
			if u, _, ok := r.BasicAuth(); ok {
				log = log.WithField("user", u)
			}
			time.Since(start)
			router.ServeHTTP(wrap, r.WithContext(logger.Set(r.Context(), log)))
			log = log.WithFields(
				"duration", time.Since(start),
				"status", http.StatusText(wrap.status),
				"statusCode", wrap.status,
				"responseSize", wrap.size,
			)
			if wrap.status == 0 {
				wrap.status = 200
			}
			if wrap.status < 400 {
				log.Info("")
			} else {
				log.Error(wrap.body.String())
			}
		}),
	}
	logger.C(ctx).Infof("starting server at %s", addr)
	if cert != "" && key != "" {
		return s.ListenAndServeTLS(cert, key)
	}
	return s.ListenAndServe()
}
