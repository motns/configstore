package main

import (
	"gopkg.in/urfave/cli.v1"
	"sort"
)

func cmdPackageLs(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	// No env provided - just list top-level envs
	if envStr == "" {
		dirs, err := ListDirs(basedir + "/env")
		if err != nil {
			return err
		}

		if len(dirs) == 0 {
			println("No environments in package")
			return nil
		}

		println("=== Environments:")
		for _, dir := range dirs {
			println(dir)
		}

		return nil
	}

	env, subenvs, err := ParseEnv(envStr, basedir, true)
	if err != nil {
		return err
	}

	// Top-level env provided - list Configstore keys
	if len(subenvs) == 0 {
		cc, err := ConfigstoreForEnv(basedir, env, subenvs, ignoreRole)
		if err != nil {
			return err
		}

		valueMap, err := cc.GetAll()
		if err != nil {
			return err
		}

		allKeys := cc.GetAllKeys()
		sort.Strings(allKeys)

		dirs, err := ListDirs(basedir + "/env/" + env)
		if err != nil {
			return err
		}

		if len(dirs) > 0 {
			println("=== Sub-environments:")
			for _, d := range dirs {
				println(d)
			}
			println("")
		}

		println("=== Configstore Values:")
		for _, k := range allKeys {
			println(k + ": " + valueMap[k])
		}

		return nil
	}

	// Sub-env provided - list overrride keys
	path, err := SubEnvPath(basedir, env, subenvs)
	if err != nil {
		return err
	}

	data, err := LoadEnvOverride(path)
	if err != nil {
		return err
	}

	dirs, err := ListDirs(path)
	if err != nil {
		return err
	}

	if len(dirs) > 0 {
		println("=== Sub-environments:")
		for _, d := range dirs {
			println(d)
		}
		println("")
	}

	allKeys := make([]string, 0)

	for k := range data {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)

	println("=== Override Values:")
	for _, k := range allKeys {
		println(k + ": " + data[k])
	}

	return nil
}
