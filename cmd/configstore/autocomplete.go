package main

import (
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
)

func ListAllPackageEnvs(basedir string, prefix string) []string {
	dirs, err := ListDirs(basedir)
	if err != nil {
		return nil
	}

	out := make([]string, 0)

	var p = ""
	if prefix != "" {
		p = prefix + "/"
	}

	for _, d := range dirs {
		out = append(out, p + d)
		out = append(out, ListAllPackageEnvs(basedir + "/" + d, p + d)...)
	}

	return out
}

type PackageCmdAutocompleteFunc func(*cli.Context, string, string)

func EnvNamesAutocomplete(c *cli.Context, basedir string, s string) {
	envs := ListAllPackageEnvs(basedir + "/env", "")

	for _, e := range envs {
		fmt.Fprintln(c.App.Writer, e)
	}
}

func EnvKeysAutocomplete(c *cli.Context, basedir string, envStr string) {
	env, err := ParseEnv(envStr, basedir, false)
	if err != nil {
		return // Do nothing
	}

	cc, err := ConfigstoreForEnv(env, true)
	if err != nil {
		return // Do nothing
	}

	for _, k := range cc.GetAllKeys("") {
		fmt.Fprintln(c.App.Writer, k)
	}
}

func ConfigstoreKeysAutocomplete(c *cli.Context) {
	cc, err := client.NewConfigstoreClient(c.String("db"), []string{}, true)
	if err != nil {
		return
	}

	for _, k := range cc.GetAllKeys("") {
		fmt.Fprintln(c.App.Writer, k)
	}
}

func PackageCmdAutocomplete(cmdAutocomplete PackageCmdAutocompleteFunc) cli.BashCompleteFunc {
	return func (c *cli.Context) {
		basedir := c.String("basedir")

		if c.NArg() == 0 {
			EnvNamesAutocomplete(c, basedir, "")
		} else if c.NArg() == 1 {
			if cmdAutocomplete == nil {
				return
			} else {
				cmdAutocomplete(c, basedir, c.Args().Get(0))
			}
		} else {
			return
		}
	}
}
