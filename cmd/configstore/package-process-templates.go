package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
)


func cmdPackageProcessTemplates(c *cli.Context) error {
	basedir := c.String("basedir")
	ignoreRole := c.Bool("ignore-role")
	envStr := c.Args().Get(0)
	outDir := c.Args().Get(1)

	env, subenvs, err := ParseEnv(envStr, basedir, true)

	if err != nil {
		return err
	}

	cc, err := ConfigstoreForEnv(basedir, env, subenvs, ignoreRole)
	if err != nil {
		return err
	}

	if !DirExists(outDir) {
		return errors.New("output directory doesn't exist: " + outDir)
	}

	templateFiles, err := ListFiles(basedir + "/template")

	if err != nil {
		return err
	}

	for _, f := range templateFiles {
		println("Processing template file: " + f)

		b, err := ioutil.ReadFile(basedir + "/template/" + f)
		if err != nil {
			return err
		}
		s := string(b)

		processed, err := cc.ProcessTemplateString(s)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(outDir + "/" + f, []byte(processed), 0644)
		if err != nil {
			return err
		}
	}

	println("Done!")
	return nil
}
