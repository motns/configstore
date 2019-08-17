package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

func cmdPackageCopy(c *cli.Context) error {
	basedir := c.String("basedir")
	srcEnvStr := c.Args().Get(0)
	destEnvStr := c.Args().Get(1)
	keyPattern := c.Args().Get(2)

	println("Loading source env: " + srcEnvStr)

	srcEnv, srcSubenvs, err := ParseEnv(srcEnvStr, basedir, true)
	if err != nil {
		return err
	}

	println("Loading destination env: " + destEnvStr)

	destEnv, destSubenvs, err := ParseEnv(destEnvStr, basedir, true)
	if err != nil {
		return err
	}

	if (len(srcSubenvs) != 0 && len(destSubenvs) == 0) || (len(srcSubenvs) == 0 && len(destSubenvs) != 0) {
		return errors.New("you can only copy values between two top-level or two sub-environments")
	}

	if len(srcSubenvs) != 0 { // We're copying between sub-environments
		srcPath, err := SubEnvPath(basedir, srcEnv, srcSubenvs)
		if err != nil {
			return err
		}

		src, err := LoadEnvOverride(srcPath)
		if err != nil {
			return err
		}

		destPath, err := SubEnvPath(basedir, destEnv, destSubenvs)
		if err != nil {
			return err
		}

		dest, err := LoadEnvOverride(destPath)
		if err != nil {
			return err
		}

		if keyPattern != "" {
			println("Copying keys matching pattern \"" + keyPattern + "\" from " + srcEnvStr + " to " + destEnvStr)
		} else {
			println("Copying keys from " + srcEnvStr + " to " + destEnvStr)
		}

		for k, v := range src {
			if keyPattern == "" || strings.Contains(k, keyPattern) {
				dest[k] = v
			}
		}

		return SaveEnvOverride(destPath, dest)

	} else { // We're copying between top-level environments
		src, err := ConfigstoreForEnv(basedir, srcEnv, srcSubenvs, false)
		if err != nil {
			return err
		}

		dest, err := ConfigstoreForEnv(basedir, destEnv, destSubenvs, false)
		if err != nil {
			return err
		}

		srcMap, err := src.GetAll(c.Bool("skip-decryption"))
		if err != nil {
			return err
		}

		if keyPattern != "" {
			println("Copying keys matching pattern \"" + keyPattern + "\" from " + srcEnvStr + " to " + destEnvStr)
		} else {
			println("Copying keys from " + srcEnvStr + " to " + destEnvStr)
		}

		for k, v := range srcMap {
			if keyPattern == "" || strings.Contains(k, keyPattern) {
				err := dest.Set(k, []byte(v.Value), v.IsSecret, v.IsBinary)
				if err != nil {
					return err
				}
			}
		}
	}

	println("Done")
	return nil
}
