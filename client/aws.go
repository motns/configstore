package client

import (
	"encoding/base64"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type AWS struct {
	sess *session.Session
}

type KMS struct {
	service *kms.KMS
}

func createAWSSession(region string, role string) (*AWS, error) {
	if region == "" {
		return nil, errors.New("region cannot be empty when setting up AWS Session")
	}

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	)

	if err != nil {
		return nil, err
	}

	// If an IAM role is set, replace Session credentials
	if role != "" {
		creds := stscreds.NewCredentials(sess, role)

		if creds == nil {
			return nil, errors.New("failed to get temporary credentials for IAM Role")
		}

		sess.Config.Credentials = creds
	}

	return &AWS{
		sess: sess,
	}, nil
}

func (a AWS) createKMS() (*KMS, error) {
	if a.sess == nil {
		return nil, errors.New("AWS Session must be initialised before services can be created")
	}

	return &KMS{
		service: kms.New(a.sess),
	}, nil
}

func (k KMS) generateDataKey(keyId string) (*kms.GenerateDataKeyOutput, error) {
	in := kms.GenerateDataKeyInput{
		KeyId:   aws.String(keyId),
		KeySpec: aws.String("AES_256"),
	}

	return k.service.GenerateDataKey(&in)
}

func (k KMS) decrypt(text []byte) ([]byte, string, error) {
	in := &kms.DecryptInput{
		CiphertextBlob: text,
	}

	out, err := k.service.Decrypt(in)
	if err != nil {
		return nil, "", err
	}

	return out.Plaintext, *out.KeyId, nil
}

func (k KMS) encrypt(keyId string, text []byte) (string, error) {
	in := &kms.EncryptInput{
		KeyId:     &keyId,
		Plaintext: text,
	}

	out, err := k.service.Encrypt(in)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out.CiphertextBlob), nil
}
