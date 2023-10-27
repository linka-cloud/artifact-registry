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
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/afero"

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

var fs = afero.NewOsFs()

func (c *client) SetupLocal(ctx context.Context, force bool) error {
	u, err := url.Parse(fmt.Sprintf("%s://%s", c.c.Options().Scheme(), c.base))
	if err != nil {
		return err
	}
	u.Path, err = url.JoinPath(u.Path, c.branch, c.repo)
	if err != nil {
		return err
	}
	if username, password, ok := c.c.Options().BasicAuth(); ok {
		u.User = url.UserPassword(username, password)
	}

	repoFile := "/etc/apk/repositories"
	// Check if the repository is already configured
	file, err := afero.ReadFile(fs, repoFile)
	lookup := fmt.Sprintf("%s%s", u.Host, u.Path)
	if err == nil && strings.Contains(string(file), lookup) && !force {
		return packages.ErrAlreadyConfigured
	}
	var lines []string
	// Remove existing repository entry
	for _, line := range strings.Split(string(file), "\n") {
		if !strings.Contains(line, lookup) {
			lines = append(lines, line)
		}
	}
	lines = append(lines, u.String())
	res, err := c.c.Get(ctx, c.path("key"))
	if err != nil {
		return fmt.Errorf("failed to get repository key: %w", err)
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
		if err := fs.Remove(fmt.Sprintf("/etc/apk/keys/%s", name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove repository key: %s", err)
		}
	}
	if name == "" {
		return fmt.Errorf("failed to get repository key: missing Content-Disposition header")
	}
	if err = afero.WriteFile(fs, name, pk, 0644); err != nil {
		return fmt.Errorf("failed to write repository key file: %w", err)
	}
	if err = afero.WriteFile(fs, repoFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write sources.list file: %w", err)
	}

	return nil
}
