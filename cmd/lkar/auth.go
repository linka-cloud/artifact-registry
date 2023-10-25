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
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var authGroup = &cobra.Group{ID: "0_auth", Title: "Authentication Commands:"}

var (
	passStdin bool

	loginCmd = &cobra.Command{
		Use:   "login [registry]",
		Short: "Login to an Artifact Registry repository",
		Example: `
Log in with username and password from command line flags:
  lkar login -u username -p password localhost:5000

Log in with username and password from stdin:
  lkar login -u username --password-stdin localhost:5000

Log in with username and password in an interactive terminal and no TLS check:
  lkar login --insecure localhost:5000
`,
		Args:    cobra.ExactArgs(1),
		GroupID: authGroup.ID,
		PreRunE: setup,
		RunE: func(cmd *cobra.Command, args []string) error {
			if user == "" {
				reader := bufio.NewReader(cmd.InOrStdin())
				cmd.Print("Username: ")
				u, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				user = strings.TrimSpace(u)
				if user == "" {
					return fmt.Errorf("username is required")
				}
			}
			if passStdin {
				reader := bufio.NewReader(cmd.InOrStdin())
				b, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				pass = strings.TrimSpace(b)
			}
			if pass == "" {
				cmd.Print("Password: ")
				b, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				fmt.Println()
				pass = strings.TrimSpace(string(b))
				if pass == "" {
					return fmt.Errorf("password is required")
				}
			}
			u := urlWithType() + "/_auth/login"
			if repository != "" {
				u = urlWithType() + fmt.Sprintf("/_auth/%s/login", repository)
			}
			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u, nil)
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
				if res.StatusCode == http.StatusUnauthorized {
					return errors.New("invalid credentials")
				}
				b, err := io.ReadAll(res.Body)
				if err != nil {
					return err
				}
				return errors.New(string(b))
			}
			if err := credsStore.Put(cmd.Context(), repoURL(), auth.Credential{Username: user, Password: pass}); err != nil {
				return err
			}
			return nil
		},
	}
	logoutCmd = &cobra.Command{
		Use:     "logout [repository]",
		Short:   "Logout from an Artifact Registry repository",
		GroupID: authGroup.ID,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := credsStore.Get(cmd.Context(), repoURL())
			if err != nil {
				return err
			}
			if creds.Username == "" && creds.Password == "" {
				return nil
			}
			return credsStore.Delete(cmd.Context(), repoURL())
		},
	}
)

func init() {
	loginCmd.Flags().BoolVar(&passStdin, "password-stdin", false, "Take the password from stdin")
	rootCmd.AddCommand(loginCmd, logoutCmd)
	rootCmd.AddGroup(authGroup)
}
