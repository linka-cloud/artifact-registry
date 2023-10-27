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

package server

import (
	"bytes"
	"io"
	"net/http"
)

func wrap(w http.ResponseWriter) *wrapWriter {
	var buf bytes.Buffer
	return &wrapWriter{ResponseWriter: w, body: &buf, w: io.MultiWriter(w, &buf)}
}

type wrapWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   *bytes.Buffer
	w      io.Writer
}

func (w *wrapWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		return 0, err
	}
	w.size += n
	return n, nil
}
