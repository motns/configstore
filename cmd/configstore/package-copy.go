package main

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

func cmdPackageCopy(c *cli.Context) error {
	basedir := c.String("basedir")
	srcEnvStr := c.Args().Get(0)
	destEnvStr := c.Args().Get(1)
	keyPattern := c.Args().Get(2)

	srcEnv, err := ParseEnv(srcEnvStr, basedir, true)
	if err != nil {
		return err
	}

	destEnv, err := ParseEnv(destEnvStr, basedir, true)
	if err != nil {
		return err
	}

	if err := cp(srcEnv, destEnv, keyPattern, c.Bool("skip-decryption"), c.Bool("recursive"), c.Bool("skip-existing")); err != nil {
		return err
	}

	fmt.Println("Done")
	return nil
}

func cp(srcEnv Env, destEnv Env, keyPattern string, skipDecryption bool, recursive bool, skipExisting bool) error {
	if (srcEnv.isSubenv() && !destEnv.isSubenv()) || (!srcEnv.isSubenv() && destEnv.isSubenv()) {
		return errors.New("you can only copy values between two top-level or two sub-environments")
	}

	if srcEnv.isSubenv() { // We're copying between two sub-environments
		if err := copySubenv(srcEnv, destEnv, keyPattern, skipExisting); err != nil {
			return err
		}
	} else { // We're copying between top-level environments
		if err := copyEnv(srcEnv, destEnv, keyPattern, skipDecryption, skipExisting); err != nil {
			return err
		}
	}

	subenvs, err := ListDirs(srcEnv.envPath())
	if err != nil {
		return err
	}

	for _, se := range subenvs {
		srcSubenv := srcEnv.getSubenv(se)
		destSubenv := destEnv.getSubenv(se)

		if !destSubenv.envExists() {
			err := CreateSubenvShared(destSubenv)
			if err != nil {
				return err
			}
		}

		err := cp(srcSubenv, destSubenv, keyPattern, skipDecryption, recursive, skipExisting)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyEnv(srcEnv Env, destEnv Env, keyPattern string, skipDecryption bool, skipExisting bool) error {
	src, err := ConfigstoreForEnv(srcEnv, false)
	if err != nil {
		return err
	}

	dest, err := ConfigstoreForEnv(destEnv, false)
	if err != nil {
		return err
	}

	srcMap, err := src.GetAll(skipDecryption)
	if err != nil {
		return err
	}

	if keyPattern != "" {
		fmt.Println("Copying keys matching pattern \"" + keyPattern + "\" from " + srcEnv.envStr() + " to " + destEnv.envStr())
	} else {
		fmt.Println("Copying keys from " + srcEnv.envStr() + " to " + destEnv.envStr())
	}

	for k, v := range srcMap {
		if keyPattern == "" || strings.Contains(k, keyPattern) {
			if !skipExisting || !dest.Exists(k) {
				err := dest.Set(k, []byte(v.Value), v.IsSecret, v.IsBinary)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func copySubenv(srcEnv Env, destEnv Env, keyPattern string, skipExisting bool) error {
	src, err := LoadEnvOverride(srcEnv.envPath())
	if err != nil {
		return err
	}

	dest, err := LoadEnvOverride(destEnv.envPath())
	if err != nil {
		return err
	}

	if keyPattern != "" {
		fmt.Println("Copying keys matching pattern \"" + keyPattern + "\" from " + srcEnv.envStr() + " to " + destEnv.envStr())
	} else {
		fmt.Println("Copying keys from " + srcEnv.envStr() + " to " + destEnv.envStr())
	}

	for k, v := range src {
		if keyPattern == "" || strings.Contains(k, keyPattern) {
			_, exists := dest[k]

			if !skipExisting || !exists {
				dest[k] = v
			}
		}
	}

	return SaveEnvOverride(destEnv.envPath(), dest)
}