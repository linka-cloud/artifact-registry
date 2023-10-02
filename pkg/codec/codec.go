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

package codec

type Codec[T any] interface {
	Encode(v T) ([]byte, error)
	Decode(b []byte) (T, error)
	Name() string
}

type CodecFuncs[T any] struct {
	Format     string
	EncodeFunc func(v T) ([]byte, error)
	DecodeFunc func(b []byte) (T, error)
}

func (c CodecFuncs[T]) Encode(a T) ([]byte, error) {
	return c.EncodeFunc(a)
}

func (c CodecFuncs[T]) Decode(b []byte) (T, error) {
	return c.DecodeFunc(b)
}

func (c CodecFuncs[T]) Name() string {
	return c.Format
}
