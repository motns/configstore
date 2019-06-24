package main

import (
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"sort"
)

func cmdLs(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.StringSlice("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	valueMap, err := cc.GetAll()
	if err != nil {
		return err
	}

	allKeys := cc.GetAllKeys()
	sort.Strings(allKeys)

	for _, k := range allKeys {
		println(k + ": " + valueMap[k])
	}

	return nil
}
