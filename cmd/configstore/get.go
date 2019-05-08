package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"github.com/motns/configstore/client"
)

func cmdGet(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.String("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	value, err := cc.Get(c.Args().Get(0))
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}
