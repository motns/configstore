package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

type AWS struct {
	sess *session.Session
}

type KMS struct {
	service *kms.KMS
}

func createAWSSession(region string) (*AWS, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	)

	if err != nil { return nil, err }

	return &AWS {
		sess: sess,
	}, nil
}

func (a AWS) createKMS() (*KMS, error) {
	if a.sess == nil {
		return nil, errors.New("AWS Session must be initialised before services can be created")
	}

	return &KMS {
		service: kms.New(a.sess),
	}, nil
}

func (k KMS) generateDataKey(keyId string) (*kms.GenerateDataKeyOutput, error) {
	in := kms.GenerateDataKeyInput {
		KeyId: aws.String(keyId),
		KeySpec: aws.String("AES_256"),
	}

	return k.service.GenerateDataKey(&in)
}

func (k KMS) decrypt(text []byte) ([]byte, error) {
	in := kms.DecryptInput{
		CiphertextBlob: text,
	}

	out, err := k.service.Decrypt(&in)
	return out.Plaintext, err
}
