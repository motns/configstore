package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
)

type ConfigstoreClient struct {
	dbFile      string
	db          ConfigstoreDB
	encryption  *Encryption
	ignoreRole  bool
	overrides   map[string]string
}


///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Private

func loadOverride(path string) (map[string]string, error) {
	var overrides = make(map[string]string)

	jsonStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonStr, &overrides); err != nil {
		return nil, err
	}

	return overrides, nil
}

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

func (c *ConfigstoreClient) ensureLatestVersion() error {
	if c.db.Version == 2 {
		return nil // Nothing to do - already latest version
	} else {
		err := c.initEncryption()
		if err != nil {
			return err
		}

		c.db.MasterKeyId = c.encryption.masterKeyId
		c.db.Version = 2
		return saveDB(c.dbFile, c.db)
	}
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
		o, exists := c.overrides[key]
		if exists {
			return o, nil
		} else {
			return entry.Value, nil
		}
	}
}

func (c *ConfigstoreClient) GetAsKMSEncrypted(key string) (string, error) {
	value, err := c.Get(key)
	if err != nil {
		return "", err
	}

	err = c.initEncryption() // Although at this point it's probably already initialised
	if err != nil {
		return "", err
	}

	encrypted, err := c.encryption.kms.encrypt(c.db.MasterKeyId, []byte(value))
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

func (c *ConfigstoreClient) GetAll() (map[string]string, error) {
	if c.dbContainsEncrypted() {
		if err := c.initEncryption(); err != nil {
			return nil, err
		}
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
			o, exists := c.overrides[k]
			if exists {
				entries[k] = o
			} else {
				entries[k] = v.Value
			}
		}
	}

	return entries, nil
}

func (c *ConfigstoreClient) GetAllKeys() []string {
	keys := make([]string, 0)

	for k := range c.db.Data {
		keys = append(keys, k)
	}

	return keys
}

func (c *ConfigstoreClient) Set(key string, rawValue []byte, isSecret bool) error {
	if key == "" {
		return errors.New("you have to specify a non-empty Key to set")
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
		return errors.New("you have to specify a non-empty Key to unset")
	}

	delete(c.db.Data, key)

	err := saveDB(c.dbFile, c.db)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigstoreClient) ProcessTemplateString(t string) (string, error) {
	tmpl, err := template.New("tmp").Parse(t)
	if err != nil {
		return "", err
	}

	templateValues, err := c.GetAll()
	if err != nil {
		return "", err
	}

	var b strings.Builder

	err = tmpl.Option("missingkey=error").Execute(&b, templateValues)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func (c *ConfigstoreClient) TestTemplateString(t string) (bool, error) {
	tmpl, err := template.New("tmp").Parse(t)
	if err != nil {
		return false, err
	}

	keys := c.GetAllKeys()
	templateValues := make(map[string]string, len(keys))
	for _, key := range keys {
		templateValues[key] = "dummy_value" // Actual value doesn't matter
	}

	err = tmpl.Option("missingkey=error").Execute(ioutil.Discard, templateValues)
	if err != nil {
		return false, err
	}

	return true, nil
}


///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Factory

func NewConfigstoreClient(dbFile string, overrideFile string, ignoreRole bool) (*ConfigstoreClient, error) {
	db, err := loadDB(dbFile)
	if err != nil {
		return nil, err
	}

	var overrides = make(map[string]string)

	if overrideFile != "" {
		overrides, err = loadOverride(overrideFile)
		if err != nil {
			return nil, err
		}

		for k := range overrides {
			mainVal, exists := db.Data[k]
			if !exists {
				return nil, errors.New("override key doesn't exist in Configstore DB: " + k)
			}

			if mainVal.IsSecret {
				return nil, errors.New("trying to override key with secret value: " + k)
			}
		}
	}

	var c = &ConfigstoreClient{
		dbFile:     dbFile,
		db:         db,
		encryption: nil,
		ignoreRole: ignoreRole,
		overrides:  overrides,
	}

	if err := c.ensureLatestVersion(); err != nil {
		return nil, err
	}

	return c, nil
}

func InitConfigstore(dir string, region string, role string, masterKey string, isInsecure bool) (*ConfigstoreClient, error) {
	var dataKey string

	if !isInsecure && masterKey == "" {
		return nil, errors.New("you have to specify --master-key if --insecure is not set")
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
		Version:    2,
		Region:     region,
		DataKey:    dataKey,
		MasterKeyId: masterKey,
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
