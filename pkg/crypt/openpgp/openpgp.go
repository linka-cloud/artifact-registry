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

package openpgp

import (
	"io"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func GenerateKeypair(name, comment, email string) (private string, public string, err error) {
	e, err := openpgp.NewEntity(name, comment, email, nil)
	if err != nil {
		return "", "", err
	}

	var priv strings.Builder
	var pub strings.Builder

	w, err := armor.Encode(&priv, openpgp.PrivateKeyType, nil)
	if err != nil {
		return "", "", err
	}
	if err := e.SerializePrivate(w, nil); err != nil {
		return "", "", err
	}
	w.Close()

	w, err = armor.Encode(&pub, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", "", err
	}
	if err := e.Serialize(w); err != nil {
		return "", "", err
	}
	w.Close()

	return priv.String(), pub.String(), nil
}

func ParseIdentity(priv string) (*openpgp.Entity, error) {
	block, err := armor.Decode(strings.NewReader(priv))
	if err != nil {
		return nil, err
	}

	return openpgp.ReadEntity(packet.NewReader(block.Body))
}

func ArmoredDetachSign(w io.Writer, priv string, message io.Reader) (err error) {
	e, err := ParseIdentity(priv)
	if err != nil {
		return err
	}
	return openpgp.ArmoredDetachSign(w, e, message, nil)
}
