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
	dbFile     string
	db         ConfigstoreDB
	encryption *Encryption
	ignoreRole bool
	overrides  map[string]string
}

///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
// Private

func loadOverrides(paths []string) (map[string]string, error) {
	var overrides = make(map[string]string)

	for _, path := range paths {
		m, err := loadOverride(path)
		if err != nil {
			return nil, fmt.Errorf("%w; Failed to load override from file: %s", err, path)
		}

		for k, v := range m {
			overrides[k] = v
		}
	}

	return overrides, nil
}

func loadOverride(path string) (map[string]string, error) {
	var overrides = make(map[string]string)

	jsonStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w; Failed to read override file", err)
	}

	if err := json.Unmarshal(jsonStr, &overrides); err != nil {
		switch t := err.(type) {
		case *json.SyntaxError:
			return nil, fmt.Errorf("%w; Failed to unmarshal json, error at position %d (\"%s\")", err, t.Offset, SafeSlice(string(jsonStr), int(t.Offset - 10), int(t.Offset + 10)))
		default:
			return nil, fmt.Errorf("%w; Failed to unmarshal json", err)
		}
	}

	return overrides, nil
}

func (c *ConfigstoreClient) initEncryption() error {
	if c.encryption == nil {
		enc, err := createEncryption(&c.db, nil, c.ignoreRole)
		if err != nil {
			return fmt.Errorf("%w; Failed to initialise encryption library", err)
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
	if c.db.Version == 1 {
		if err := c.migrateToV2(); err != nil {
			return err
		}
	}

	if c.db.Version == 2 {
		if err := c.migrateToV3(); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigstoreClient) migrateToV2() error {
	err := c.initEncryption()
	if err != nil {
		return err
	}

	c.db.MasterKeyId = c.encryption.masterKeyId
	c.db.Version = 2
	return saveDB(c.dbFile, c.db)
}

func (c *ConfigstoreClient) migrateToV3() error {
	c.db.Version = 3
	// New attributes will be filled in via zero values
	return saveDB(c.dbFile, c.db)
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
			return "", fmt.Errorf("%w; Failed to decrypt value for key: %s", err, key)
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

func (c *ConfigstoreClient) Exists(key string) bool {
	if key == "" {
		return false
	}

	_, exists := c.db.Data[key]
	return exists
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

func (c *ConfigstoreClient) GetAll(skipDecryption bool) (map[string]ConfigstoreDBValue, error) {
	if c.dbContainsEncrypted() && !skipDecryption {
		if err := c.initEncryption(); err != nil {
			return nil, err
		}
	}

	entries := make(map[string]ConfigstoreDBValue, len(c.db.Data))

	for k, v := range c.db.Data {
		if v.IsSecret {
			var value string

			if skipDecryption {
				value = "(secret)"
			} else {
				decoded, err := c.encryption.decrypt(v.Value)
				if err != nil {
					return nil, fmt.Errorf("%w; Failed to decrypt value for key: %s", err, k)
				}
				value = decoded
			}

			entries[k] = ConfigstoreDBValue{
				Value:    value,
				IsSecret: v.IsSecret,
				IsBinary: v.IsBinary,
			}
		} else {
			o, exists := c.overrides[k]
			if exists {
				entries[k] = ConfigstoreDBValue{
					Value:    o,
					IsSecret: v.IsSecret,
					IsBinary: v.IsBinary,
				}
			} else {
				entries[k] = v
			}
		}
	}

	return entries, nil
}

func (c *ConfigstoreClient) GetAllValues(skipDecryption bool) (map[string]string, error) {
	entries, err := c.GetAll(skipDecryption)
	if err != nil {
		return nil, err
	}

	valueMap := make(map[string]string, 0)

	for k, v := range entries {
		valueMap[k] = v.Value
	}

	return valueMap, nil
}

func (c *ConfigstoreClient) GetAllKeys(keyFilter string) []string {
	keys := make([]string, 0)

	for k := range c.db.Data {
		if keyFilter != "" {
			if strings.Contains(k, keyFilter) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}

	return keys
}

func (c *ConfigstoreClient) Set(key string, rawValue []byte, isSecret bool, isBinary bool) error {
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

		value = encrypted
	} else {
		value = string(rawValue)
	}

	c.db.Data[key] = ConfigstoreDBValue{
		Value:    value,
		IsSecret: isSecret,
		IsBinary: isBinary,
	}

	err := saveDB(c.dbFile, c.db)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigstoreClient) Encrypt(key string) error {
	if key == "" {
		return errors.New("you have to specify a non-empty Key to encrypt")
	}

	entry, exists := c.db.Data[key]
	if !exists {
		return errors.New("key does not exist in Configstore: " + key)
	}

	// Already encrypted - leave alone
	if entry.IsSecret {
		return nil
	}

	err := c.initEncryption()
	if err != nil {
		return err
	}

	encrypted, err := c.encryption.encrypt([]byte(entry.Value))
	if err != nil {
		return err
	}

	c.db.Data[key] = ConfigstoreDBValue{
		Value:    string(encrypted),
		IsSecret: true,
		IsBinary: entry.IsBinary,
	}

	err = saveDB(c.dbFile, c.db)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigstoreClient) Decrypt(key string) error {
	if key == "" {
		return errors.New("you have to specify a non-empty Key to decrypt")
	}

	entry, exists := c.db.Data[key]
	if !exists {
		return errors.New("key does not exist in Configstore: " + key)
	}

	// Already plain text - leave alone
	if !entry.IsSecret {
		return nil
	}

	err := c.initEncryption()
	if err != nil {
		return err
	}

	decrypted, err := c.encryption.decrypt(entry.Value)
	if err != nil {
		return fmt.Errorf("%w; Failed to decrypt value for key: %s", err, key)
	}

	c.db.Data[key] = ConfigstoreDBValue{
		Value:    decrypted,
		IsSecret: false,
		IsBinary: entry.IsBinary,
	}

	err = saveDB(c.dbFile, c.db)
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

	templateValues, err := c.GetAllValues(false)
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

	keys := c.GetAllKeys("")
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

func NewConfigstoreClient(dbFile string, overrideFiles []string, ignoreRole bool) (*ConfigstoreClient, error) {
	db, err := loadDB(dbFile)
	if err != nil {
		return nil, err
	}

	var overrides = make(map[string]string)

	if len(overrideFiles) != 0 {
		overrides, err = loadOverrides(overrideFiles)
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
		Version:     2,
		Region:      region,
		DataKey:     dataKey,
		MasterKeyId: masterKey,
		IsInsecure:  isInsecure,
		Role:        role,
		Data:        make(map[string]ConfigstoreDBValue),
	}

	dbFile := dir + "/configstore.json"
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
