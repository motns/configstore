package main

import (
	"errors"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageSet(c *cli.Context) error {
	basedir := c.String("basedir")

	envStr := c.Args().Get(0)
	env, subenv, err := ParseEnv(envStr, basedir)

	if err != nil {
		return err
	}

	if subenv == "" { // We're updating a main environment
		dbFile := basedir + "/env/" + env + "/configstore.json"

		cc, err := client.NewConfigstoreClient(dbFile, "", c.Bool("ignore-role"))
		if err != nil {
			return err
		}

		isSecret := c.Bool("secret")
		key := c.Args().Get(1)
		val := c.Args().Get(2)

		return SetCmdShared(cc, isSecret, key, val)

	} else { // We're updating a sub-environment
		overrides, err := LoadEnvOverride(basedir, env, subenv)
		if err != nil {
			return err
		}

		if isSecret := c.Bool("secret"); isSecret == true {
			return errors.New("secret values cannot be stored in overrides")
		}

		key := c.Args().Get(1)
		val := c.Args().Get(2)

		overrides[key] = val

		return SaveEnvOverride(basedir, env, subenv, overrides)
	}
}
