package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Encryption struct {
	dataKey     []byte
	masterKeyId string
	kms         *KMS
}

// Used to create an Encryption object for encrypting/decrypting secrets. Will initialise
// an AWS API Session, and create a KMS instance if one is not passed in.
// If an IAM Role was defined when the Configstore was created, the `ignoreRole` flag can
// be used to ignore (not assume) that IAM Role, and instead use the default credentials - this
// is useful for example on EC2 servers, which cannot assume regular IAM roles, and have to rely
// on Instance Roles instead (you do however have to make sure that the Instance Role has access
// to the KMS Key used for the Configstore)
func createEncryption(db *ConfigstoreDB, kms *KMS, ignoreRole bool) (*Encryption, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(db.DataKey)
	if err != nil {
		return nil, fmt.Errorf("%w; Failed to load ciphertext", err)
	}

	var dataKey []byte
	var masterKey = ""

	if db.IsInsecure {
		// The dataKey is stored as plain text
		dataKey = ciphertext
	} else {
		if kms == nil {
			role := db.Role

			if ignoreRole == true {
				role = ""
			}

			aws, err := createAWSSession(db.Region, role)
			if err != nil {
				return nil, fmt.Errorf("%w; Failed to initialise AWS Session", err)
			}

			kms, err = aws.createKMS()
			if err != nil {
				return nil, fmt.Errorf("%w; Failed to initialise KMS", err)
			}
		}

		dataKey, masterKey, err = kms.decrypt(ciphertext)
		if err != nil {
			return nil, fmt.Errorf("%w; Failed to decrypt Data Key", err)
		}
	}

	return &Encryption{
		dataKey:     dataKey,
		masterKeyId: masterKey,
		kms:         kms,
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
