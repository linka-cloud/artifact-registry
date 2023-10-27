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

package registry

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func DebugTransport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	return &debugTransport{t: t}
}

type debugTransport struct {
	t http.RoundTripper
}

func (s *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", bytes)
	resp, err := s.t.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	bytes, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", bytes)

	return resp, err
}
