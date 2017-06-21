package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func cmdInit(c *cli.Context) error {
	dir := c.String("dir")
	region := c.String("region")
	role := c.String("role")
	masterKey := c.String("master-key")
	isInsecure := c.Bool("insecure")

	// Check that destination folder exists
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	if !isInsecure && masterKey == "" {
		return errors.New("You have to specify --master-key if --insecure is not set")
	}

	var dataKey string

	if isInsecure {
		fmt.Printf("Initialising **Insecure** Configstore into directory: %s\n", dir)
		// Since we're storing it as plain text, it doesn't really matter anyway
		dataKey = "OfvuQJ0Cis1CvnFV2KTTYv3WCPKXOIord3OBDc0kwcU="
		region = ""
		role = ""
	} else {
		if role != "" {
			fmt.Printf("Initialising Configstore for Region \"%s\" with Master Key \"%s\", using IAM Role \"%s\", into directory: %s\n", region, masterKey, role, dir)
		} else {
			fmt.Printf("Initialising Configstore for Region \"%s\" with Master Key \"%s\" into directory: %s\n", region, masterKey, dir)
		}

		aws, err := createAWSSession(region, role)
		if err != nil {
			return err
		}

		kms, _ := aws.createKMS()

		generated, err := kms.generateDataKey(masterKey)
		if err != nil {
			return err
		}
		dataKey = base64.StdEncoding.EncodeToString(generated.CiphertextBlob)
	}

	db := ConfigstoreDB{
		Version:    1,
		Region:     region,
		DataKey:    dataKey,
		IsInsecure: isInsecure,
		Role:       role,
		Data:       make(map[string]ConfigstoreDBValue),
	}

	if err := saveConfigStore(dir+"/configstore.json", db); err != nil {
		return err
	}

	return nil
}
