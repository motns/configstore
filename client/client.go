package client

import (
	"encoding/base64"
	"errors"
	"fmt"
)

type ConfigstoreClient struct {
	dbFile      string
	db          ConfigstoreDB
	encryption  *Encryption
	ignoreRole  bool
}


///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Private

func (c *ConfigstoreClient) initEncryption() error {
	if c.encryption == nil {
		enc, err := createEncryption(&c.db, nil, c.ignoreRole)
		if err != nil {
			return err
		}

		c.encryption = enc
		return nil
	} else {
		return nil
	}
}

func (c ConfigstoreClient) dbContainsEncrypted() bool {
	for _, v := range c.db.Data {
		if v.IsSecret {
			return true
		}
	}

	return false
}


///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Public

func (c *ConfigstoreClient) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("you have to specify a non-empty Key to get")
	}

	entry, exists := c.db.Data[key]
	if !exists {
		return "", errors.New("key does not exist in Configstore: " + key)
	}

	if entry.IsSecret {
		err := c.initEncryption()
		if err != nil {
			return "", err
		}

		decrypted, err := c.encryption.decrypt(entry.Value)
		if err != nil {
			return "", err
		}

		return decrypted, nil
	} else {
		return entry.Value, nil
	}
}

func (c *ConfigstoreClient) GetAll() (map[string]string, error) {
	if c.dbContainsEncrypted() {
		c.initEncryption()
	}

	entries := make(map[string]string, len(c.db.Data))

	for k, v := range c.db.Data {
		if v.IsSecret {
			decoded, err := c.encryption.decrypt(v.Value)
			if err != nil {
				return nil, err
			}

			entries[k] = decoded
		} else {
			entries[k] = v.Value
		}
	}

	return entries, nil
}

func (c *ConfigstoreClient) Set(key string, rawValue []byte, isSecret bool) error {
	if key == "" {
		return errors.New("You have to specify a non-empty Key to set")
	}

	var value string

	if isSecret {
		err := c.initEncryption()
		if err != nil {
			return err
		}

		encrypted, err := c.encryption.encrypt(rawValue)
		if err != nil {
			return err
		}

		value = string(encrypted)
	} else {
		value = string(rawValue)
	}

	c.db.Data[key] = ConfigstoreDBValue{
		Value: value,
		IsSecret: isSecret,
	}

	err := saveDB(c.dbFile, c.db)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigstoreClient) Unset(key string) error {
	if key == "" {
		return errors.New("You have to specify a non-empty Key to unset")
	}

	delete(c.db.Data, key)

	err := saveDB(c.dbFile, c.db)
	if err != nil {
		return err
	}

	return nil
}


///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Factory

func NewConfigstoreClient(dbFile string, ignoreRole bool) (*ConfigstoreClient, error) {
	db, err := loadDB(dbFile)
	if err != nil {
		return nil, err
	}

	return &ConfigstoreClient{
		dbFile:     dbFile,
		db:         db,
		encryption: nil,
		ignoreRole: ignoreRole,
	}, nil
}

func InitConfigstore(dir string, region string, role string, masterKey string, isInsecure bool) (*ConfigstoreClient, error) {
	var dataKey string

	if !isInsecure && masterKey == "" {
		return nil, errors.New("You have to specify --master-key if --insecure is not set")
	}

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
			return nil, err
		}

		kms, _ := aws.createKMS()

		generated, err := kms.generateDataKey(masterKey)
		if err != nil {
			return nil, err
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

	dbFile := dir+"/configstore.json"
	err := saveDB(dbFile, db)
	if err != nil {
		return nil, err
	}

	return &ConfigstoreClient{
		dbFile:     dbFile,
		db:         db,
		encryption: nil,
		ignoreRole: false, // There's no reason we'd want to do this right after initialisation
	}, nil
}
