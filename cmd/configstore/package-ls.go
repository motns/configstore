package main

import (
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)


func cmdPackageLs(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	var env string

	// If no environment is defined, we'll gather all "main" environments, load all their keys, and build
	// a table with the keys as rows and the environments as columns.
	// Missing keys, and keys with a values that differs between environments will be highlighted.
	if envStr != "" {
		var err error
		env, _, err = ParseEnv(envStr, basedir, true)

		if err != nil {
			return err
		}
	} else {
		env = ""
	}

	if env == "" {
		envs, err := ListDirs(basedir + "/env")

		if err != nil {
			return err
		}

		sort.Strings(envs)

		configstores := make(map[string]*client.ConfigstoreClient)

		for _, env := range envs {
			println("Loading Configstore: " + env)
			cc, err := ConfigstoreForEnv(basedir, env, nil, ignoreRole)

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

		for k := range keySet {
			allKeys = append(allKeys, k)
		}

		sort.Strings(allKeys)

		RenderTable(allKeys, allValues, envs, false)

	} else { // List sub-environments under the specified main environment *only*
		cc, err := ConfigstoreForEnv(basedir, env, nil, ignoreRole)

		if err != nil {
			return err
		}

		keySet := make(map[string]int)
		ccValues, err := cc.GetAll()
		ccKeys := cc.GetAllKeys()

		if err != nil {
			return err
		}

		for _, k := range ccKeys {
			keySet[k] = 0
		}

		var subenvs []string

		if _, err = os.Stat(basedir + "/env/" + env + "/subenv"); !os.IsNotExist(err) {
			subenvs, err = ListDirs(basedir + "/env/" + env + "/subenv")

			if err != nil {
				return err
			}
		}

		sort.Strings(subenvs)

		allValues := make(map[string]map[string]string)
		allValues["(default)"] = make(map[string]string)

		for _, se := range subenvs {
			path, err := SubEnvPath(basedir, env, []string{se})
			if err != nil {
				return err
			}

			data, err := LoadEnvOverride(path)

			if err != nil {
				return err
			}

			allValues[se] = data

			for k := range data {
				keySet[k] = 0
			}
		}

		allKeys := make([]string, 0)

		for k := range keySet {
			allKeys = append(allKeys, k)
			allValues["(default)"][k] = ccValues[k]
		}

		sort.Strings(allKeys)

		// Prepend this for display purposes only
		subenvs = append([]string{"(default)"}, subenvs...)

		RenderTable(allKeys, allValues, subenvs, true)
	}

	return nil
}
