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

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type data interface {
	string | []byte
}

func toBytes[T data](d T) []byte {
	switch d := any(d).(type) {
	case string:
		return []byte(d)
	case []byte:
		return d
	default:
		panic("unexpected data type")
	}
}

func Encrypt[T, U data](key T, content U) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	kb, cb := toBytes(key), toBytes(content)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(kb[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data using aesGCM.Seal
	// Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	return aesGCM.Seal(nonce, nonce, cb, nil), nil
}

func Decrypt[T, U data](key T, enc U) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	kb, encb := toBytes(key), toBytes(enc)
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(kb[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce, ciphertext := encb[:nonceSize], encb[nonceSize:]

	// Decrypt the data
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
