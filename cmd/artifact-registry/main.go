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
	"crypto/sha256"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.linka.cloud/artifact-registry/pkg/packages"
)

const (
	EnvAddr    = "ARTIFACT_REGISTRY_ADDRESS"
	EnvBackend = "ARTIFACT_REGISTRY_BACKEND"
	EnvKey     = "ARTIFACT_REGISTRY_KEY"
)

var (
	addr = ":9887"
	// backend = "192.168.10.11:5000"
	backend = "docker.io"

	key = ""

	cmd = &cobra.Command{
		Use:          "artifact-registry",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd.Context()); err != nil {
				logrus.Fatal(err)
			}
		},
	}
)

func main() {
	cmd.Flags().StringVar(&addr, "addr", envDefault(EnvAddr, addr), "address to listen on [$"+EnvAddr+"]")
	cmd.Flags().StringVar(&backend, "backend", envDefault(EnvBackend, backend), "registry backend [$"+EnvBackend+"]")
	cmd.Flags().StringVar(&key, "key", envDefault(EnvKey, key), "key to encrypt the repositories keys [$"+EnvKey+"]")
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

type wrapWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *wrapWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}
	w.size += n
	return n, nil
}

func run(ctx context.Context) error {
	if key == "" {
		return errors.New("key must be set")
	}
	logrus.Infof("intializing artifact registry using backend %s", backend)
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrap := &wrapWriter{ResponseWriter: w}
			start := time.Now()
			next.ServeHTTP(wrap, r)
			time.Since(start)
			status := wrap.status
			if status == 0 {
				status = 200
			}
			logrus.WithFields(logrus.Fields{
				"method":    r.Method,
				"path":      r.URL.Path,
				"remote":    r.RemoteAddr,
				"duration":  time.Since(start),
				"status":    status,
				"size":      wrap.size,
				"userAgent": r.UserAgent(),
			}).Infof("")
		})
	})
	h := sha256.New()
	h.Write([]byte(key))
	if err := packages.Init(ctx, r, backend, h.Sum(nil)); err != nil {
		return err
	}
	logrus.Infof("starting server at %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		return err
	}
	return nil
}

func envDefault(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
