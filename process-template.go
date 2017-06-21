package main

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"text/template"
)

func cmdProcessTemplate(c *cli.Context) error {
	dbFile := c.String("db")
	db, err := loadConfigstore(dbFile)
	if err != nil {
		return err
	}

	templateFilePath := c.Args().Get(0)
	if templateFilePath == "" {
		return errors.New("You have to specify the path to a Go template file as the first argument")
	}

	enc, err := createEncryption(&db, nil, c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	/////////////////////////////////////////////////////////////////////////////
	// Load values from Configstore

	templateValues := make(map[string]string)

	for key, entry := range db.Data {
		if entry.IsSecret {
			decrypted, err := enc.decrypt(entry.Value)
			if err != nil {
				return err
			}

			templateValues[key] = decrypted
		} else {
			templateValues[key] = entry.Value
		}
	}

	/////////////////////////////////////////////////////////////////////////////
	// Load and execute template, and print to Stdout

	tmpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return err
	}

	return tmpl.Option("missingkey=error").Execute(os.Stdout, templateValues)
}
