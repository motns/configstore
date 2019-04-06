package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"sort"
	"github.com/CultBeauty/configstore/client"
)

func cmdLs(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.String("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	valueMap, err := cc.GetAll()
	if err != nil {
		return err
	}

	entries := make([]string, 0, len(valueMap))

	for k, v := range valueMap {
		entries = append(entries, k+": "+v)
	}

	sort.Strings(entries)

	for _, v := range entries {
		fmt.Println(v)
	}

	return nil
}
