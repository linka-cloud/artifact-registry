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

package helm

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed setup.sh
var script string

var (
	scriptTemplate = template.Must(template.New("setup.sh").Parse(script))
)

type SetupArgs struct {
	User     string
	Password string
	Scheme   string
	Host     string
	Path     string
	Name     string
}

func (c *client) SetupLocal(ctx context.Context, force bool) error {
	if _, err := exec.LookPath("helm"); err != nil {
		return err
	}
	var name string
	if c.repository != "" {
		name = strings.NewReplacer("/", "-").Replace(c.repository)
	} else {
		name = strings.NewReplacer("/", "-", ".", "-").Replace(strings.TrimPrefix(strings.Split(c.c.Options().Host(), ":")[0], Name+"."))
	}
	as := []string{"repo", "add", name, fmt.Sprintf("%s://%s", c.c.Options().Scheme(), c.base)}
	user, pass, _ := c.c.Options().BasicAuth()
	if user != "" {
		as = append(as, "--username", user)
	}
	if pass != "" {
		as = append(as, "--password", pass)
	}
	if force {
		as = append(as, "--force-update")
	}
	if _, err := exec.LookPath("helm"); err != nil {
		return err
	}
	b, err := exec.CommandContext(ctx, "helm", as...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(b))
	}
	return nil
}
