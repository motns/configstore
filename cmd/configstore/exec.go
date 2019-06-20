package main

import (
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"os/exec"
	"strings"
)

func cmdExec(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.StringSlice("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	cmdTemplate := strings.Join(c.Args(), " ")

	cmdStr, err := cc.ProcessTemplateString(cmdTemplate)

	if err != nil {
		return err
	}

	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)

	out, err := cmd.Output()

	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}
