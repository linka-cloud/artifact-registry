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
	"fmt"
	"os/exec"
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
	if _, err := exec.LookPath("helm"); err != nil {
		return err
	}
	as := []string{"repo", "add", args.Name, args.Scheme + "://" + args.Host + args.Path}
	if args.User != "" {
		as = append(as, "--username", args.User)
	}
	if args.Password != "" {
		as = append(as, "--password", args.Password)
	}
	if force {
		as = append(as, "--force-update")
	}
	b, err := exec.CommandContext(ctx, "helm", as...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(b))
	}
	return nil
}
