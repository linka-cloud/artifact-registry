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

package packages

import (
	"context"
	"io"
)

type Client interface {
	Signer
	Setuper
	Puller
	Pusher
	Deleter
}

type Puller interface {
	Pull(ctx context.Context, path string) (io.ReadCloser, int64, error)
}

type Pusher interface {
	Push(ctx context.Context, r io.Reader) error
}

type Deleter interface {
	Delete(ctx context.Context, path string) error
}

type Signer interface {
	Key(ctx context.Context) (string, error)
}

type Setuper interface {
	SetupScript(ctx context.Context) (string, error)
	SetupLocal(ctx context.Context, force bool) error
}
