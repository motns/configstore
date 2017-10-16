package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"github.com/CultBeauty/configstore/client"
)

func cmdInit(c *cli.Context) error {
	dir := c.String("dir")
	region := c.String("region")
	role := c.String("role")
	masterKey := c.String("master-key")
	isInsecure := c.Bool("insecure")

	// Check that destination folder exists
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	if !isInsecure && masterKey == "" {
		return errors.New("You have to specify --master-key if --insecure is not set")
	}

	_, err := client.InitConfigstore(dir, region, role, masterKey, isInsecure)
	if err != nil {
		return err
	}

	return nil
}
