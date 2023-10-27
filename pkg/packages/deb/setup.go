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
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"

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

var fs = afero.NewOsFs()

func (c *client) SetupLocal(ctx context.Context, force bool) error {
	u, err := url.Parse(fmt.Sprintf("%s://%s", c.c.Options().Scheme(), c.base))
	if err != nil {
		return err
	}

	var name string
	if c.repository != "" {
		name = strings.NewReplacer("/", "-").Replace(c.repository)
	} else {
		name = strings.NewReplacer("/", "-", ".", "-").Replace(strings.TrimPrefix(strings.Split(u.Host, ":")[0], Name+"."))
	}

	u.Path, err = url.JoinPath(u.Path, c.distribution, c.component)
	if err != nil {
		return err
	}

	rs := filepath.Join("/etc/apt/sources.list.d", name+".list")
	if _, err := fs.Stat(rs); err == nil && !force {
		return packages.ErrAlreadyConfigured
	}

	// Pull repository key
	pub, err := c.Key(ctx)
	if err != nil {
		return fmt.Errorf("failed to download repository key: %w", err)
	}

	k := filepath.Join("/etc/apt/trusted.gpg.d", name+".asc")
	if err := afero.WriteFile(fs, k, []byte(pub), 0644); err != nil {
		return fmt.Errorf("failed to write repository key file: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if user, pass, ok := c.c.Options().BasicAuth(); ok {
		authConfig := fmt.Sprintf("machine %s login %s password %s", u.String(), user, pass)
		authFile := filepath.Join("/etc/apt/auth.conf.d", name+".conf")
		if err := afero.WriteFile(fs, authFile, []byte(authConfig), 0644); err != nil {
			return fmt.Errorf("failed to write auth config file: %w", err)
		}
	}

	// Add repository to sources.list
	s := fmt.Sprintf("deb %s://%s %s %s", c.c.Options().Scheme(), c.base, c.distribution, c.component)
	if err := afero.WriteFile(fs, rs, []byte(s), 0644); err != nil {
		return fmt.Errorf("failed to write sources.list file: %w", err)
	}
	return nil
}
