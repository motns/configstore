package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"github.com/CultBeauty/configstore/client"
)

func cmdUnset(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("You have to specify a Key to unset as the first argument")
	}

	err = cc.Unset(key)
	if err != nil {
		return err
	}

	return nil
}