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

package cache

import (
	"time"
)

var (
	// DefaultTTL is the default TTL for cache items.
	// It matches the default TTL of docker.io registry.
	DefaultTTL = 6 * time.Hour
	d          = New()
)

func Set(key string, value any, opts ...Option) {
	d.Set(key, value, opts...)
}
func Get(key string) (any, bool) {
	return d.Get(key)
}
