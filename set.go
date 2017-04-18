package main

import (
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"gopkg.in/urfave/cli.v1"
)

func cmdSet(c *cli.Context) error {
	dbFile := c.String("db")
	db, err := loadConfigstore(dbFile)
	if err != nil {
		return err
	}

	isSecret := c.Bool("secret")

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("You have to specify a Key to set as the first argument")
	}

	var value string

	if isSecret {
		enc, err := createEncryption(&db, nil)
		if err != nil {
			return err
		}

		fmt.Print("Secret:")

		pass, err := gopass.GetPasswd()
		if err != nil {
			return err
		}

		encrypted, err := enc.encrypt(pass)
		if err != nil {
			return err
		}

		value = string(encrypted)
	} else {
		value = c.Args().Get(1)
	}

	db.Data[key] = ConfigstoreDBValue{
		IsSecret: isSecret,
		Value:    value,
	}

	if err := saveConfigStore(dbFile, db); err != nil {
		return err
	}

	return nil
}
