package main

import (
	"errors"
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
)

func cmdProcessTemplate(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), c.String("override"), c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	templateFilePath := c.Args().Get(0)
	if templateFilePath == "" {
		return errors.New("you have to specify the path to a Go template file as the first argument")
	}

	b, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return err
	}
	s := string(b)

	out, err := cc.ProcessTemplateString(s)
	if err != nil {
		return err
	}

	_, err = fmt.Println(out)
	if err != nil {
		return err
	}

	return nil
}
