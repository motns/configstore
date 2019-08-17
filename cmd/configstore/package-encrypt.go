package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageEncrypt(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	env, err := ParseEnv(envStr, basedir, true)
	if err != nil {
		return err
	}

	if env.isSubenv() {
		return errors.New("encrypt command not supported for sub-environment")
	}

	cc, err := ConfigstoreForEnv(env, ignoreRole)
	if err != nil {
		return err
	}

	err = cc.Encrypt(c.Args().Get(1))
	if err != nil {
		return err
	}

	return nil
}
