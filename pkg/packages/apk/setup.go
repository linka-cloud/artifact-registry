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

package apk

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"go.linka.cloud/artifact-registry/pkg/packages"
)

//go:embed setup.sh
var script string

var scriptTemplate = template.Must(template.New("setup.sh").Parse(script))

type SetupArgs struct {
	User       string
	Password   string
	Scheme     string
	Host       string
	Path       string
	Branch     string
	Repository string
}

func Setup(ctx context.Context, args SetupArgs, force bool) error {
	username := args.User
	password := args.Password
	scheme := args.Scheme
	repoHost := args.Host
	repoPath := args.Path
	branch := args.Branch
	repository := args.Repository

	repoURL := fmt.Sprintf("%s://%s%s/%s/%s", scheme, repoHost, repoPath, branch, repository)
	repoPattern := fmt.Sprintf("%s%s/%s/%s", repoHost, repoPath, branch, repository)
	repoFile := "/etc/apk/repositories"

	// Check if the repository is already configured
	file, err := os.ReadFile(repoFile)
	if err == nil && strings.Contains(string(file), repoPattern) && !force {
		return packages.ErrAlreadyConfigured
	}

	if username != "" {
		repoURL = fmt.Sprintf("%s://%s:%s@%s%s/%s/%s", scheme, username, password, repoHost, repoPath, branch, repository)
	}

	// Remove existing repository entry
	lines := []string{repoURL}
	for _, line := range strings.Split(string(file), "\n") {
		if !strings.Contains(line, repoPattern) {
			lines = append(lines, line)
		}
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/key", repoURL), nil)
	if err != nil {
		return fmt.Errorf("failed to get repository key: %w", err)
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to get repository key: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to get repository key: %s", string(b))
	}
	defer res.Body.Close()

	pk, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read repository key data: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var name string
	if h := res.Header.Get("Content-Disposition"); h != "" {
		name = fmt.Sprintf("/etc/apk/keys/%s", strings.TrimPrefix(h, "attachment; filename="))
		if err := os.Remove(fmt.Sprintf("/etc/apk/keys/%s", name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove repository key: %s", err)
		}
	}
	if name == "" {
		return fmt.Errorf("failed to get repository key: missing Content-Disposition header")
	}
	if err = os.WriteFile(name, pk, 0644); err != nil {
		return fmt.Errorf("failed to write repository key file: %w", err)
	}
	if err = os.WriteFile(repoFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write sources.list file: %w", err)
	}

	return nil
}
