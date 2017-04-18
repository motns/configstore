package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
)

func cmdUnset(c *cli.Context) error {
	dbFile := c.String("db")
	db, err := loadConfigstore(dbFile)
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("You have to specify a Key to unset as the first argument")
	}

	delete(db.Data, key)

	if err := saveConfigStore(dbFile, db); err != nil {
		return err
	}

	return nil
}
