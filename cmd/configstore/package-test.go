package main

import (
	"errors"
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
)


func cmdPackageTest(c *cli.Context) error {
	basedir := c.String("basedir")

	envs, err := ListDirs(basedir + "/env")

	if err != nil {
		return err
	}

	if len(envs) == 0 {
		return errors.New("no environments in package")
	}

	path1 := basedir + "/env/" + envs[0] + "/configstore.json"
	cc1, err := client.NewConfigstoreClient(path1, make([]string, 0), true)

	// Only run key comparison if we have more than one Configstores
	if len(envs) != 1 {
		if err != nil {
			return err
		}

		for _, env := range envs[1:] {
			path2 := basedir + "/env/" + env + "/configstore.json"
			cc2, err := client.NewConfigstoreClient(path2, make([]string, 0), true)

			if err != nil {
				return err
			}

			fmt.Printf("Comparing keys for \"%s\" and \"%s\"\n", path1, path2)
			out, err := CompareKeys(cc1, cc2)

			if err != nil {
				PrintLines(out)
				return err
			}
		}
	}

	templateFiles, err := ListFiles(basedir + "/template")

	if err != nil {
		return err
	}

	for _, f := range templateFiles {
		println("Testing template file: " + f)

		b, err := ioutil.ReadFile(basedir + "/template/" + f)
		if err != nil {
			return err
		}
		s := string(b)

		_, err = cc1.TestTemplateString(s)

		if err != nil {
			return err
		}
	}

	println("All tests passed!")
	return nil
}
