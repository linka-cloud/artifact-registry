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
