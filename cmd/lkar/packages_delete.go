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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func newPkgDeleteCmd(typ string) *cobra.Command {
	return &cobra.Command{
		Use:     fmt.Sprintf("delete [repository] [path]"),
		Short:   fmt.Sprintf("Delete %s package from the repository", typ),
		Aliases: []string{"rm", "remove", "del"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlWithType(typ)+"/"+repository+"/"+args[1], nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(user, pass)
			res, err := client().Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				b, err := io.ReadAll(res.Body)
				if err != nil {
					return err
				}
				return errors.New(string(b))
			}
			return nil
		},
	}
}
