package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func cmdPackageGet(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	env, subenvs, err := ParseEnv(envStr, basedir, true)

	if err != nil {
		return err
	}

	cc, err := ConfigstoreForEnv(basedir, env, subenvs, ignoreRole)
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
