package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"github.com/motns/configstore/client"
)

func cmdUnset(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, make([]string, 0), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("you have to specify a Key to unset as the first argument")
	}

	err = cc.Unset(key)
	if err != nil {
		return err
	}

	return nil
}
