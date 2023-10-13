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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

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

func Setup(ctx context.Context, args SetupArgs, force bool) error {
	// Extract input values from the struct
	user := args.User
	password := args.Password
	scheme := args.Scheme
	repoHost := args.Host
	repoPath := args.Path
	repoName := filepath.Base(repoPath)

	// Set the repository URL based on user and password
	var repoURL string
	if user != "" {
		repoURL = fmt.Sprintf("%s://%s:%s@%s%s", scheme, user, password, repoHost, repoPath)
	} else {
		repoURL = fmt.Sprintf("%s://%s%s", scheme, repoHost, repoPath)
	}

	// Check if the repository file already exists
	f := filepath.Join("/etc/yum.repos.d", repoName+".repo")
	if _, err := os.Stat(f); err == nil && !force {
		return packages.ErrAlreadyConfigured
	}

	// Determine the package manager to use (dnf or yum)
	var prog string
	if _, err := exec.LookPath("dnf"); err == nil {
		prog = "dnf"
	} else {
		prog = "yum"
	}

	// Check if the package manager supports config-manager
	var hasConfigManager bool
	if prog == "dnf" {
		cmd := exec.Command(prog, "config-manager", "--help")
		if cmd.Run() == nil {
			hasConfigManager = true
		}
	}

	// Check if curl is available for dnf-based systems
	if prog == "dnf" && hasConfigManager {
		res, err := http.Get(fmt.Sprintf("%s.repo", repoURL))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(res.Body)
			return fmt.Errorf("failed to download Repository file: %s", string(b))
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if err := os.WriteFile(f, b, 0644); err != nil {
			return err
		}
	} else {
		// Use dnf config-manager or yum to add the repository
		cmd := exec.CommandContext(ctx, prog, "config-manager", "--add-repo", repoURL+".repo")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
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
