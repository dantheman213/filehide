package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(key, data []byte) []byte {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	encryptedPayload := aesgcm.Seal(nil, nonce, data, nil)
	return encryptedPayload
}

func Decrypt(key, encryptedData []byte) []byte {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	decryptedPayload, err := aesgcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		panic(err.Error())
	}

	return decryptedPayload
}