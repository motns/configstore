package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageUnset(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")
	key := c.Args().Get(1)

	env, err := ParseEnv(envStr, basedir, true)
	if err != nil {
		return err
	}

	if !env.envExists() {
		return errors.New("env doesn't exist: " + env.envStr())
	}

	if env.isMainEnv() {
		cc, err := ConfigstoreForEnv(env, ignoreRole)
		if err != nil {
			return err
		}

		if err := cc.Unset(key); err != nil {
			return err
		}
	} else {
		overrides, err := LoadEnvOverride(env.envPath())
		if err != nil {
			return err
		}

		delete(overrides, key)
		if err := SaveEnvOverride(env.envPath(), overrides); err != nil {
			return err
		}
	}

	return nil
}
