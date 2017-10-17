package main

import (
	"fmt"
	"errors"
	"io/ioutil"
	"gopkg.in/urfave/cli.v1"
	"text/template"
	"github.com/CultBeauty/configstore/client"
)

func cmdTestTemplate(c *cli.Context) error {
	dbFile := c.String("db")

	cc, err := client.NewConfigstoreClient(dbFile, c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	templateFilePath := c.Args().Get(0)
	if templateFilePath == "" {
		return errors.New("You have to specify the path to a Go template file as the first argument")
	}

	keys := cc.GetAllKeys()
	templateValues := make(map[string]string, len(keys))
	for _, key := range keys {
		templateValues[key] = "dummy_value" // Actual value doesn't matter
	}

	// Check if template can be parsed
	tmpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	err = tmpl.Option("missingkey=error").Execute(ioutil.Discard, templateValues)
	if err != nil {
		return err
	}

	fmt.Println("OK")

	return nil
}
