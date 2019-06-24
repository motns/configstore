package main

import (
	"errors"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageSet(c *cli.Context) error {
	basedir := c.String("basedir")

	envStr := c.Args().Get(0)
	env, subenvs, err := ParseEnv(envStr, basedir, true)

	if err != nil {
		return err
	}

	if len(subenvs) == 0 { // We're updating a main environment
		dbFile := basedir + "/env/" + env + "/configstore.json"

		cc, err := client.NewConfigstoreClient(dbFile, make([]string, 0), c.Bool("ignore-role"))
		if err != nil {
			return err
		}

		isSecret := c.Bool("secret")
		key := c.Args().Get(1)
		val := c.Args().Get(2)

		return SetCmdShared(cc, isSecret, key, val)

	} else { // We're updating a sub-environment
		path, err := SubEnvPath(basedir, env, subenvs)
		if err != nil {
			return err
		}

		overrides, err := LoadEnvOverride(path)
		if err != nil {
			return err
		}

		if isSecret := c.Bool("secret"); isSecret == true {
			return errors.New("secret values cannot be stored in overrides")
		}

		key := c.Args().Get(1)
		val := c.Args().Get(2)

		overrides[key] = val

		return SaveEnvOverride(path, overrides)
	}
}
