package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageEncrypt(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	env, subenvs, err := ParseEnv(envStr, basedir, true)

	if err != nil {
		return err
	}

	if len(subenvs) != 0 {
		return errors.New("encrypt command not supported for sub-environment")
	}

	cc, err := ConfigstoreForEnv(basedir, env, subenvs, ignoreRole)
	if err != nil {
		return err
	}

	err = cc.Encrypt(c.Args().Get(1))
	if err != nil {
		return err
	}

	return nil
}
