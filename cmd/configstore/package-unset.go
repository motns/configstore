package main

import (
	"gopkg.in/urfave/cli.v1"
)


func cmdPackageUnset(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")
	key := c.Args().Get(1)

	env, subenvs, err := ParseEnv(envStr, basedir, true)

	if err != nil {
		return err
	}

	if len(subenvs) == 0 {
		cc, err := ConfigstoreForEnv(basedir, env, subenvs, ignoreRole)
		if err != nil {
			return err
		}

		if err := cc.Unset(key); err != nil {
			return err
		}
	} else {
		path, err := SubEnvPath(basedir, env, subenvs)
		if err != nil {
			return err
		}

		overrides, err := LoadEnvOverride(path)
		if err != nil {
			return err
		}

		delete(overrides, key)
		if err := SaveEnvOverride(path, overrides); err != nil {
			return err
		}
	}

	return nil
}
