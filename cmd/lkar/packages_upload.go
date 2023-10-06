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
	"os"
	"path"

	"github.com/spf13/cobra"
)

func newPkgUploadCmd(typ string) *cobra.Command {
	use := fmt.Sprintf("upload [repository] [path]")
	index := 1
	upload := func(args []string) string {
		return "/upload"
	}
	switch typ {
	case "apk":
		use = fmt.Sprintf("upload [repository] [branch] [apk-repository] [path]")
		index = 3
		upload = func(args []string) string {
			return path.Join("/", args[1], args[2], "upload")
		}
	case "deb":
		use = fmt.Sprintf("upload [repository] [distribution] [component] [path]")
		index = 3
		upload = func(args []string) string {
			return path.Join("/", args[1], args[2], "upload")
		}
	}
	return &cobra.Command{
		Use:     use,
		Short:   fmt.Sprintf("Upload %s package to the repository", typ),
		Aliases: []string{"up", "put", "create"},
		Args:    cobra.ExactArgs(index + 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			f, err := os.Open(args[index])
			if err != nil {
				return err
			}
			defer f.Close()
			i, err := f.Stat()
			if err != nil {
				return err
			}
			pw := newProgressReader(f, i.Size())
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, url(typ)+"/"+repository+upload(args), pw)
			if err != nil {
				return err
			}
			req.SetBasicAuth(user, pass)
			req.ContentLength = i.Size()
			go pw.Run(ctx)
			res, err := client().Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusCreated {
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
