package main

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"sort"
)

func cmdPackageLs(c *cli.Context) error {
	basedir := c.String("basedir")
	envStr := c.Args().Get(0)
	ignoreRole := c.Bool("ignore-role")

	// No envName provided - just list top-level envs
	if envStr == "" {
		dirs, err := ListDirs(basedir + "/env")
		if err != nil {
			return err
		}

		if len(dirs) == 0 {
			fmt.Println("No environments in package")
			return nil
		}

		fmt.Println("=== Environments:")
		for _, dir := range dirs {
			fmt.Println(dir)
		}

		return nil
	}

	env, err := ParseEnv(envStr, basedir, true)
	if err != nil {
		return err
	}

	// Top-level envName provided - list Configstore keys
	if env.isMainEnv() {
		cc, err := ConfigstoreForEnv(env, ignoreRole)
		if err != nil {
			return err
		}

		entries, err := cc.GetAll(c.Bool("skip-decryption"))
		if err != nil {
			return err
		}

		allKeys := cc.GetAllKeys()
		sort.Strings(allKeys)

		dirs, err := ListDirs(env.mainEnvPath())
		if err != nil {
			return err
		}

		if len(dirs) > 0 {
			fmt.Println("=== Sub-environments:")
			for _, d := range dirs {
				fmt.Println(d)
			}
			fmt.Println("")
		}

		fmt.Println("=== Configstore Values:")
		for _, k := range allKeys {
			e := entries[k]

			if e.IsBinary {
				fmt.Println(k + ": (binary)")
			} else {
				fmt.Println(k + ": " + e.Value)
			}
		}

		return nil
	}

	// Subenv provided - list override keys
	if !env.envExists() {
		return errors.New("sub-environment doesn't exist: " + env.envStr())
	}

	data, err := LoadEnvOverride(env.envPath())
	if err != nil {
		return err
	}

	dirs, err := ListDirs(env.envPath())
	if err != nil {
		return err
	}

	if len(dirs) > 0 {
		fmt.Println("=== Sub-environments:")
		for _, d := range dirs {
			fmt.Println(d)
		}
		fmt.Println("")
	}

	allKeys := make([]string, 0)

	for k := range data {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)

	fmt.Println("=== Override Values:")
	for _, k := range allKeys {
		fmt.Println(k + ": " + data[k])
	}

	return nil
}
