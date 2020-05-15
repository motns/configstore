package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

func cmdPackageEnvs(c *cli.Context) error {
	basedir := c.String("basedir")
	return printDirs(basedir + "/env", 0)
}

func printDirs(basedir string, indent int) error {
	dirs, err := ListDirs(basedir)
	if err != nil {
		return fmt.Errorf("%w; Failed to list top-level directories at: %s", err, basedir)
	}

	for _, dir := range dirs {
		out := ""
		if indent == 0 {
			out = formatGreen(dir)
		} else if indent == 1 {
			out = formatCyan(dir)
		} else {
			out = dir
		}

		fmt.Println(strings.Repeat(" ", indent * 2) + "/" + out)

		if err := printDirs(basedir + "/" + dir, indent + 1); err != nil {
			return err
		}
	}

	return nil
}
