package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"github.com/CultBeauty/configstore/client"
)

func cmdAsKMSEnc(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, c.String("override"), c.Bool("ignore-role"))
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
