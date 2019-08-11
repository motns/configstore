package main

import (
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdGet(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.StringSlice("override"), c.Bool("ignore-role"))
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
