package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageSet(c *cli.Context) error {
	basedir := c.String("basedir")

	envStr := c.Args().Get(0)
	env, err := ParseEnv(envStr, basedir, true)
	if err != nil {
		return err
	}

	if !env.envExists() {
		return errors.New("env doesn't exist: " + env.envStr())
	}

	isSecret := c.Bool("secret")
	isBinary := c.Bool("binary")
	key := c.Args().Get(1)
	val := c.Args().Get(2)

	rawValue, err := ReadRawValue(isSecret, val)
	if err != nil {
		return err
	}

	if env.isMainEnv() { // We're updating a main environment
		cc, err := ConfigstoreForEnv(env, c.Bool("ignore-role"))
		if err != nil {
			return err
		}

		return cc.Set(key, rawValue, isSecret, isBinary)
	} else { // We're updating a sub-environment
		overrides, err := LoadEnvOverride(env.envPath())
		if err != nil {
			return err
		}

		if isSecret == true {
			return errors.New("secret values cannot be stored in overrides")
		}

		overrides[key] = string(rawValue)

		return SaveEnvOverride(env.envPath(), overrides)
	}
}
