package main

import (
	"errors"
	"fmt"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
)

func cmdTestTemplate(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, make([]string, 0), c.Bool("ignore-role"))
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

	_, err = cc.TestTemplateString(s)
	if err != nil {
		return err
	}

	fmt.Println("OK")

	return nil
}
