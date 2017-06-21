package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "configstore"
	app.HelpName = "configstore"
	app.Usage = "Manage plain-text and encrypted credentials, using local JSON file as storage"
	app.UsageText = "configstore [global options] command [command options]"
	app.EnableBashCompletion = true
	app.Version = "0.0.6"

	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Initialise a new Configstore",
			Action: cmdInit,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir",
					Usage: "The directory to create the Configstore file in",
					Value: ".",
				},
				cli.StringFlag{
					Name:  "region",
					Usage: "The AWS Region the KMS key will be created in",
					Value: "eu-west-1",
				},
				cli.StringFlag{
					Name:  "role",
					Usage: "The IAM Role to assume before executing AWS API operations",
				},
				cli.StringFlag{
					Name:  "master-key",
					Usage: "The name of the AWS KMS key to be used as the master encryption key",
				},
				cli.BoolFlag{
					Name:  "insecure",
					Usage: "Initialise this Configstore with a plain-text encryption key (not backed by KMS)",
				},
			},
		},
		{
			Name:      "set",
			Usage:     "Set a new value, or update an existing one in the Configstore",
			ArgsUsage: "key value",
			Action:    cmdSet,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.BoolFlag{
					Name:  "secret",
					Usage: "Whether this value is sensitive (to be encrypted)",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
		{
			Name:      "get",
			Usage:     "Get a value from the Configstore",
			ArgsUsage: "key",
			Action:    cmdGet,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
		{
			Name:   "ls",
			Usage:  "List all keys and their respective values from the Configstore",
			Action: cmdLs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
		{
			Name:      "unset",
			Usage:     "Remove a value from the Configstore",
			ArgsUsage: "key",
			Action:    cmdUnset,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
			},
		},
		{
			Name:      "process_template",
			Usage:     "Takes a GO template file, and fills in values from this Configstore",
			ArgsUsage: "/path/to/template",
			Action:    cmdProcessTemplate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
	}

	app.Run(os.Args)
}
