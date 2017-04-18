package main

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func cmdGet(c *cli.Context) error {
	dbFile := c.String("db")
	db, err := loadConfigstore(dbFile)
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("You have to specify a Key to get as the first argument")
	}

	entry, exists := db.Data[key]
	if !exists {
		return errors.New("Key does not exist in Configstore: " + key)
	}

	if entry.IsSecret {
		enc, err := createEncryption(&db, nil)
		if err != nil {
			return err
		}

		decrypted, err := enc.decrypt(entry.Value)
		if err != nil {
			return err
		}

		fmt.Println(decrypted)
	} else {
		fmt.Println(entry.Value)
	}

	return nil
}
