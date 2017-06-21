package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// Types

type ConfigstoreDB struct {
	Version    int                           `json:"version"`
	Region     string                        `json:"region"`
	Role       string                        `json:"role"`
	IsInsecure bool                          `json:"is_insecure"`
	DataKey    string                        `json:"data_key"`
	Data       map[string]ConfigstoreDBValue `json:"data"`
}

func (c ConfigstoreDB) validate() (ConfigstoreDB, error) {
	if c.Version == 0 {
		return c, errors.New("Missing key in Configstore DB: version")
	}

	if c.DataKey == "" {
		return c, errors.New("Missing key in Configstore DB: data_key")
	}

	if c.Data == nil {
		return c, errors.New("Missing key in Configstore DB: data")
	}

	return c, nil
}

type ConfigstoreDBValue struct {
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// Helpers

// loadConfigstore looks for a file at path dbFile, loads the JSON string
// from it, and parses it into a ConfigStoreDB. An error is returned if
// the file is missing, is not valid JSON, or the parsed contents
// are missing required fields.
func loadConfigstore(dbFile string) (ConfigstoreDB, error) {
	var db ConfigstoreDB

	if dbFile == "" {
		return ConfigstoreDB{}, errors.New("Cannot load Configstore DB from empty path!")
	}

	jsonStr, err := ioutil.ReadFile(dbFile)
	if err != nil {
		return ConfigstoreDB{}, err
	}

	if err := json.Unmarshal(jsonStr, &db); err != nil {
		return ConfigstoreDB{}, err
	}

	return db.validate()
}

// saveConfigStore takes the provided ConfigstoreDB, marshals it into pretty-printed
// JSON, and then writes said JSON string into a file specified by dbFile
func saveConfigStore(dbFile string, db ConfigstoreDB) error {
	jsonStr, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return errors.New("Failed to marshal Configstore DB into JSON")
	}

	if err := ioutil.WriteFile(dbFile, jsonStr, 0644); err != nil {
		return err
	}

	return nil
}
