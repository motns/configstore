package main

import (
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdSet(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), make([]string, 0), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	key := c.Args().Get(0)
	val := c.Args().Get(1)
	isSecret := c.Bool("secret")
	isBinary := c.Bool("binary")

	return SetCmdShared(cc, isSecret, isBinary, key, val)
}
