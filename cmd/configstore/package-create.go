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
	env, subenv, err := ParseEnv(envStr, "")

	if err != nil {
		return err
	}

	println("Creating environment: " + envStr)

	if subenv == "" { // We're creating a main environment
		if EnvExists(basedir, env) {
			return errors.New("environment already exists: " + env)
		}

		dir := basedir + "/env/" + env

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
		if EnvExists(basedir, env) == false {
			return errors.New("main environment doesn't exists: " + env)
		}

		if SubEnvExists(basedir, env, subenv) {
			return errors.New("sub-environment already exists: " + envStr)
		}

		dir := basedir + "/env/" + env + "/subenv/" + subenv

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(dir + "/override.json", os.O_CREATE|os.O_RDWR, 0666)

		if err != nil {
			return err
		}

		if _, err := f.WriteString("{}"); err != nil {
			return err
		}

		if err = f.Close(); err != nil {
			return err
		}
	}

	return nil
}
