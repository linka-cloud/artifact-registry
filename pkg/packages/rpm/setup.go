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

package rpm

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"

	"go.linka.cloud/artifact-registry/pkg/packages"
)

//go:embed setup.sh
var script string

var (
	scriptTemplate = template.Must(template.New("setup.sh").Parse(script))
	repoTemplate   = template.Must(template.New("repo").Parse(`[{{.Name}}]
name={{.Name}}
baseurl={{.URL}}
enabled=1
gpgcheck=1
gpgkey={{.URL}}/{{.Key}}
{{- if .User }}
username={{.User}}
password={{.Password}}
{{- end }}
`))
)

type SetupArgs struct {
	User     string
	Password string
	Scheme   string
	Host     string
	Path     string
	Name     string
}

var fs = afero.NewOsFs()

func (c *client) SetupLocal(ctx context.Context, force bool) error {
	u, err := url.Parse(fmt.Sprintf("%s://%s", c.c.Options().Scheme(), c.base))
	if err != nil {
		return err
	}

	var name string
	if c.repo != "" {
		name = strings.NewReplacer("/", "-").Replace(c.repo)
	} else {
		name = strings.NewReplacer("/", "-", ".", "-").Replace(strings.TrimPrefix(strings.Split(u.Host, ":")[0], Name+"."))
	}

	if user, pass, ok := c.c.Options().BasicAuth(); ok {
		u.User = url.UserPassword(user, pass)
	}

	// Check if the repository file already exists
	f := filepath.Join("/etc/yum.repos.d", name+".repo")
	if _, err := fs.Stat(f); err == nil && !force {
		return packages.ErrAlreadyConfigured
	}

	def, err := c.Repo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get repository definition: %w", err)
	}
	if err := afero.WriteFile(fs, f, []byte(def), 0600); err != nil {
		return err
	}
	return nil
}

func repoDefinition(w io.Writer, name, url, key, user, password string) error {
	data := map[string]string{
		"Name":     name,
		"URL":      url,
		"Key":      key,
		"User":     user,
		"Password": password,
	}
	return repoTemplate.ExecuteTemplate(w, "repo", data)
}
