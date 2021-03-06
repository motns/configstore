package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// Types

type ConfigstoreDB struct {
	Version     int                           `json:"version"`
	Region      string                        `json:"region"`
	Role        string                        `json:"role"`
	IsInsecure  bool                          `json:"is_insecure"`
	DataKey     string                        `json:"data_key"`
	MasterKeyId string                        `json:"master_key_id,omitempty"`
	Data        map[string]ConfigstoreDBValue `json:"data"`
}

type ConfigstoreDBValue struct {
	Value    string `json:"value"`
	IsBinary bool   `json:"is_binary"`
	IsSecret bool   `json:"is_secret"`
}

func (c ConfigstoreDB) validate() (ConfigstoreDB, error) {
	if c.Version == 0 {
		return c, errors.New("missing key in Configstore DB: version")
	}

	if c.DataKey == "" {
		return c, errors.New("missing key in Configstore DB: data_key")
	}

	if c.Data == nil {
		return c, errors.New("missing key in Configstore DB: data")
	}

	return c, nil
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// Helpers

// loadDB looks for a file at path dbFile, loads the JSON string
// from it, and parses it into a ConfigStoreDB. An error is returned if
// the file is missing, is not valid JSON, or the parsed contents
// are missing required fields.
func loadDB(dbFile string) (ConfigstoreDB, error) {
	var db ConfigstoreDB

	if dbFile == "" {
		return ConfigstoreDB{}, errors.New("cannot load Configstore DB from empty path")
	}

	jsonStr, err := ioutil.ReadFile(dbFile)
	if err != nil {
		return ConfigstoreDB{}, fmt.Errorf("%w; Failed to load DB file: %s", err, dbFile)
	}

	if err := json.Unmarshal(jsonStr, &db); err != nil {
		switch t := err.(type) {
		case *json.SyntaxError:
			return ConfigstoreDB{}, fmt.Errorf("%w; Failed to unmarshal json from DB file \"%s\", error at position %d (\"%s\")", err, dbFile, t.Offset, SafeSlice(string(jsonStr), int(t.Offset - 10), int(t.Offset + 10)))
		default:
			return ConfigstoreDB{}, fmt.Errorf("%w; Failed to unmarshal json from DB file \"%s\"", err, dbFile)
		}
	}

	return db.validate()
}

// saveDB takes the provided ConfigstoreDB, marshals it into pretty-printed
// JSON, and then writes said JSON string into a file specified by dbFile
func saveDB(dbFile string, db ConfigstoreDB) error {
	jsonStr, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return errors.New("failed to marshal Configstore DB into JSON")
	}

	if err := ioutil.WriteFile(dbFile, jsonStr, 0644); err != nil {
		return err
	}

	return nil
}
