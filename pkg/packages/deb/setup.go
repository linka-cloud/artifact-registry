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

package deb

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"go.linka.cloud/artifact-registry/pkg/packages"
)

//go:embed setup.sh
var script string

var scriptTemplate = template.Must(template.New("setup.sh").Parse(script))

type SetupArgs struct {
	User      string
	Password  string
	Scheme    string
	Host      string
	Path      string
	Name      string
	Dist      string
	Component string
}

func Setup(ctx context.Context, args SetupArgs, force bool) error {
	rs := filepath.Join("/etc/apt/sources.list.d", args.Name+".list")
	if _, err := os.Stat(rs); err == nil && !force {
		return packages.ErrAlreadyConfigured
	}

	repoURL := fmt.Sprintf("%s://%s%s", args.Scheme, args.Host, args.Path)

	var repoAuth string
	if args.User != "" {
		repoAuth = fmt.Sprintf("%s://%s:%s@%s%s", args.Scheme, args.User, args.Password, args.Host, args.Path)
	} else {
		repoAuth = fmt.Sprintf("%s://%s%s/%s/%s", args.Scheme, args.Host, args.Path, args.Dist, args.Component)
	}

	// Download repository key
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/repository.key", repoAuth), nil)
	if err != nil {
		return fmt.Errorf("failed to create repository key request: %w", err)
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to download repository key: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to download Repository key: %s", string(b))
	}

	pk, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read repository key data: %w", err)
	}

	k := filepath.Join("/etc/apt/trusted.gpg.d", args.Name+".asc")
	if err := os.WriteFile(k, pk, 0644); err != nil {
		return fmt.Errorf("failed to write repository key file: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if args.User != "" {
		authConfig := fmt.Sprintf("machine %s login %s password %s", repoURL, args.User, args.Password)
		authFile := filepath.Join("/etc/apt/auth.conf.d", args.Name+".conf")
		if err := os.WriteFile(authFile, []byte(authConfig), 0644); err != nil {
			return fmt.Errorf("failed to write auth config file: %w", err)
		}
	}

	// Add repository to sources.list
	s := fmt.Sprintf("deb %s %s %s", repoURL, args.Dist, args.Component)
	if err := os.WriteFile(rs, []byte(s), 0644); err != nil {
		return fmt.Errorf("failed to write sources.list file: %w", err)
	}
	return nil
}
