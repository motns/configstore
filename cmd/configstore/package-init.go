package main

import (
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
)

func cmdPackageInit(c *cli.Context) error {
	basedir := c.Args().Get(0)

	if basedir == "" {
		basedir = "./config"
	}

	println("Initialising environment structure in basedir: " + basedir)

	if err := os.MkdirAll(basedir + "/env", 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(basedir + "/template", 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(basedir + "/env/.gitkeep", []byte(""), 0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile(basedir + "/template/.gitkeep", []byte(""), 0644); err != nil {
		return err
	}

	return nil
}
