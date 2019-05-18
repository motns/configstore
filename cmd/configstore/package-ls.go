package main

import (
	"github.com/motns/configstore/client"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"sort"
)


func cmdPackageLs(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	var env string

	if envStr != "" {
		var err error
		env, _, err = ParseEnv(envStr, basedir)

		if err != nil {
			return err
		}
	} else {
		env = ""
	}

	if env == "" {
		envs := make([]string, 0)
		entries, err := ioutil.ReadDir(basedir + "/env")

		if err != nil {
			return err
		}

		for _, e := range entries {
			if e.IsDir() {
				envs = append(envs, e.Name())
			}
		}

		sort.Strings(envs)

		configstores := make(map[string]*client.ConfigstoreClient)

		for _, env := range envs {
			cc, err := ConfigstoreForEnv(basedir, env, "", ignoreRole)

			if err != nil {
				return err
			}

			configstores[env] = cc
		}

		keySet := make(map[string]int)
		allValues := make(map[string]map[string]string)

		for env, cc := range configstores {
			data, err := cc.GetAll()

			if err != nil {
				return err
			}

			allValues[env] = data

			for _, k := range cc.GetAllKeys() {
				keySet[k] = 0
			}
		}

		allKeys := make([]string, 0)

		for k, _ := range keySet {
			allKeys = append(allKeys, k)
		}

		sort.Strings(allKeys)
		headers := append([]string{"Key / Env"}, envs...)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(headers)

		for _, k := range allKeys {
			cols := make([]string, 0)
			cols = append(cols, k)

			for _, e := range envs {
				cols = append(cols, allValues[e][k])
			}

			table.Append(cols)
		}

		table.Render()
	} else {

	}

	return nil
}
