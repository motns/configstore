package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type Encryption struct {
	dataKey []byte
}

func createEncryption(db *ConfigstoreDB, kms *KMS) (*Encryption, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(db.DataKey)
	if err != nil {
		return nil, err
	}

	var key []byte

	if db.IsInsecure {
		// The key is stored as plain text
		key = ciphertext
	} else {
		if kms == nil {
			aws, err := createAWSSession(db.Region)
			if err != nil {
				return nil, err
			}
			kms, _ = aws.createKMS()
		}

		key, err = kms.decrypt(ciphertext)
		if err != nil {
			return nil, err
		}
	}

	return &Encryption{
		dataKey: key,
	}, nil
}

func (e Encryption) encrypt(text []byte) (string, error) {
	block, err := aes.NewCipher(e.dataKey)
	if err != nil {
		return "", err
	}

	b := base64.StdEncoding.EncodeToString(text)

	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

func (e Encryption) decrypt(encoded string) (string, error) {
	text, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.dataKey)
	if err != nil {
		return "", err
	}

	if len(text) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return "", err
	}

	return string(data), nil
}
