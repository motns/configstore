package main

import (
	"errors"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func cmdPackageCreate(c *cli.Context) error {
	basedir := c.String("basedir")

	envStr := c.Args().Get(0)
	env, err := ParseEnv(envStr, basedir, false)
	if err != nil {
		return err
	}

	println("Creating environment: " + envStr)

	if env.isMainEnv() { // We're creating a main environment
		if env.envExists() {
			return errors.New("environment already exists: " + env.envStr())
		}

		dir := basedir + "/env/" + env.envName

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		region := c.String("region")
		role := c.String("role")
		masterKey := c.String("master-key")
		isInsecure := c.Bool("insecure")

		if !isInsecure && masterKey == "" {
			return errors.New("you have to specify --master-key if --insecure is not set")
		}

		_, err := client.InitConfigstore(dir, region, role, masterKey, isInsecure)
		if err != nil {
			return err
		}

	} else { // We're creating a sub-environment
		err := CreateSubenvShared(env)
		if err != nil {
			return err
		}
	}

	return nil
}
