package main

import (
	"errors"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)


func cmdCompareKeys(c *cli.Context) error {
	dbFile1 := c.Args().Get(0)
	if dbFile1 == "" {
		return errors.New("you have to provide a Configstore DB as the first argument")
	}

	dbFile2 := c.Args().Get(1)
	if dbFile2 == "" {
		return errors.New("you have to provide a Configstore DB to compare against as the second argument")
	}

	cc1, err := client.NewConfigstoreClient(dbFile1, make([]string, 0), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	cc2, err := client.NewConfigstoreClient(dbFile2, make([]string, 0), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	out, err := CompareKeys(cc1, cc2)
	PrintLines(out)
	return err
}
