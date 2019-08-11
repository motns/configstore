package main

import (
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdDecrypt(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.StringSlice("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	err = cc.Decrypt(c.Args().Get(0))
	if err != nil {
		return err
	}

	return nil
}
