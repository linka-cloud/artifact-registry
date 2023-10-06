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
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.linka.cloud/artifact-registry/pkg/logger"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/repository"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const (
	EnvAddr    = "ARTIFACT_REGISTRY_ADDRESS"
	EnvBackend = "ARTIFACT_REGISTRY_BACKEND"
	EnvKey     = "ARTIFACT_REGISTRY_AES_KEY"
	EnvDomain  = "ARTIFACT_REGISTRY_DOMAIN"
)

var (
	addr = ":9887"
	// backend = "192.168.10.11:5000"
	backend = "docker.io"

	domain = ""

	aesKey = ""

	noHTTPS = false

	insecure = false

	tagPerArtifact = false

	key, cert string

	clientCA string

	cmd = &cobra.Command{
		Use:          "artifact-registry",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO(adphi): validate host
			opts := []storage.Option{storage.WithHost(backend)}
			if noHTTPS {
				opts = append(opts, storage.WithPlainHTTP())
			}
			if insecure {
				opts = append(opts, storage.WithInsecure())
			}
			if tagPerArtifact {
				opts = append(opts, storage.WithArtifactTags())
			}
			if clientCA != "" {
				p := x509.NewCertPool()
				b, err := os.ReadFile(clientCA)
				if err != nil {
					logrus.Fatal(err)
				}
				if !p.AppendCertsFromPEM(b) {
					logrus.Fatal(err)
				}
				opts = append(opts, storage.WithClientCA(p))
			}
			if err := run(cmd.Context(), opts...); err != nil {
				logrus.Fatal(err)
			}
		},
	}
)

func main() {
	cmd.Flags().StringVar(&addr, "addr", envDefault(EnvAddr, addr), "address to listen on [$"+EnvAddr+"]")
	cmd.Flags().StringVar(&backend, "backend", envDefault(EnvBackend, backend), "registry backend [$"+EnvBackend+"]")
	cmd.Flags().StringVar(&aesKey, "aes-key", envDefault(EnvKey, aesKey), "AES key to encrypt the repositories keys [$"+EnvKey+"]")
	cmd.Flags().StringVar(&domain, "domain", envDefault(EnvDomain, domain), "domain to use to serve the repositories as subdomains [$"+EnvDomain+"]")
	cmd.Flags().BoolVar(&noHTTPS, "no-https", noHTTPS, "disable backend registry client https")
	cmd.Flags().BoolVar(&insecure, "insecure", insecure, "disable backend registry client tls verification")
	cmd.Flags().BoolVar(&tagPerArtifact, "tag-artifacts", tagPerArtifact, "tag artifacts manifests")
	cmd.Flags().StringVar(&clientCA, "client-ca", "", "tls client certificate authority")
	cmd.Flags().StringVar(&cert, "cert", "", "tls certificate")
	cmd.Flags().StringVar(&key, "key", "", "tls key")
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
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

func run(ctx context.Context, opts ...storage.Option) error {
	if aesKey == "" {
		return fmt.Errorf("environment variable $%s must be set", EnvKey)
	}
	logrus.Infof("intializing artifact registry using backend %s", backend)
	router := mux.NewRouter()
	k := sha256.Sum256([]byte(aesKey))
	// TODO(adphi): client tls
	ctx = storage.WithOptions(ctx, append(opts, storage.WithKey(k[:]))...)
	if err := repository.Init(ctx, router, domain); err != nil {
		return err
	}
	if err := packages.Init(ctx, router, domain); err != nil {
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
			log := logrus.WithFields(logrus.Fields{
				"method":    r.Method,
				"path":      r.URL.Path,
				"remote":    r.RemoteAddr,
				"userAgent": r.UserAgent(),
			})
			if u, _, ok := r.BasicAuth(); ok {
				log = log.WithField("user", u)
			}
			time.Since(start)
			router.ServeHTTP(wrap, r.WithContext(logger.Set(r.Context(), log)))
			log = log.WithFields(logrus.Fields{
				"duration":     time.Since(start),
				"status":       http.StatusText(wrap.status),
				"statusCode":   wrap.status,
				"responseSize": wrap.size,
			})
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
	logrus.Infof("starting server at %s", addr)
	if cert != "" && key != "" {
		return s.ListenAndServeTLS(cert, key)
	}
	return s.ListenAndServe()
}

func envDefault(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
