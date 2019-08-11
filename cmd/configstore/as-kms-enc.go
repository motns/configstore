package main

import (
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdAsKMSEnc(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, c.StringSlice("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	value, err := cc.GetAsKMSEncrypted(c.Args().Get(0))
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}
