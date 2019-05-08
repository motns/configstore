package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"text/template"
	"github.com/motns/configstore/client"
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

	templateValues, err := cc.GetAll()
	if err != nil {
		return err
	}

	// Load and execute template, and print to Stdout
	tmpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	return tmpl.Option("missingkey=error").Execute(os.Stdout, templateValues)
}
