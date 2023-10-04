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

package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func PublicKeyAndFingerprintFromPrivateKey(priv string) (pub []byte, fp []byte, err error) {
	privPem, _ := pem.Decode([]byte(priv))
	if privPem == nil {
		return nil, nil, fmt.Errorf("failed to decode private key pem")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, nil, err
	}
	fp, err = PublicKeyFingerprint(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubPem, err := pubPem(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return []byte(pubPem), fp, nil
}

// PublicKeyFingerprint creates a fingerprint of the given key.
// The fingerprint is the sha256 sum of the PKIX structure of the key.
func PublicKeyFingerprint(key crypto.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}

	checksum := sha256.Sum256(bytes)

	return checksum[:], nil
}

func GenerateKeyPair() (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}
	priv, err := privPem(key)
	if err != nil {
		return "", "", err
	}
	pub, err := pubPem(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	return priv, pub, nil
}

func privPem(priv *rsa.PrivateKey) (string, error) {
	b := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	return string(b), nil
}

func pubPem(pub *rsa.PublicKey) (string, error) {
	b, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	b = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	})
	return string(b), nil
}
