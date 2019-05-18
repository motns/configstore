package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageGet(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	env, subenv, err := ParseEnv(envStr, basedir)

	if err != nil {
		return err
	}

	cc, err := ConfigstoreForEnv(basedir, env, subenv, ignoreRole)
	if err != nil {
		return err
	}

	value, err := cc.Get(c.Args().Get(1))
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}
