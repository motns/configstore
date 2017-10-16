package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"text/template"
	"github.com/CultBeauty/configstore/client"
)

func cmdProcessTemplate(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	templateFilePath := c.Args().Get(0)
	if templateFilePath == "" {
		return errors.New("You have to specify the path to a Go template file as the first argument")
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
