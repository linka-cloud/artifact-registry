// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rpm

import (
	"bytes"
	"io"

	"github.com/sassoftware/go-rpmutils"
	"golang.org/x/crypto/openpgp"

	"go.linka.cloud/artifact-registry/pkg/buffer"
)

func SignPackage(rpm *buffer.HashedBuffer, privateKey string) (reader io.Reader, signSize int64, original int64, err error) {
	// TODO(adphi): check if we can use openpgp.ParseIdentity instead
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(privateKey)))
	if err != nil {
		// failed to parse  key
		return nil, 0, 0, err
	}
	entity := keyring[0]
	h, err := rpmutils.SignRpmStream(rpm, entity.PrivateKey, nil)
	if err != nil {
		// error signing rpm
		return nil, 0, 0, err
	}
	signBlob, err := h.DumpSignatureHeader(false)
	if err != nil {
		// error writing sig header
		return nil, 0, 0, err
	}
	if len(signBlob)%8 != 0 {
		return nil, 0, 0, err
	}
	return bytes.NewReader(signBlob), int64(len(signBlob)), int64(h.OriginalSignatureHeaderSize()), nil
}
