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
	app.Version = "2.6.0"

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
					Name:  "binary",
					Usage: "Indicate whether this value contains binary data (instead of plain text)",
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
				cli.StringSliceFlag{
					Name:  "override",
					Usage: "JSON file with key-value pairs for overriding non-secret values in Configstore DB",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
			BashComplete: ConfigstoreKeysAutocomplete,
		},
		{
			Name:      "encrypt",
			Usage:     "Take an existing value from the DB and change it to be a secret (encrypted) value",
			ArgsUsage: "key",
			Action:    cmdEncrypt,
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
			BashComplete: ConfigstoreKeysAutocomplete,
		},
		{
			Name:      "decrypt",
			Usage:     "Take an existing value from the given environment and change it to be a plain (unencrypted) value",
			ArgsUsage: "key",
			Action:    cmdDecrypt,
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
			BashComplete: ConfigstoreKeysAutocomplete,
		},
		{
			Name:      "as_kms_enc",
			Usage:     "Retrieve a value from the Configstore, encrypt it with the KMS Master Key for this DB, and return the result",
			ArgsUsage: "key",
			Action:    cmdAsKMSEnc,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.StringSliceFlag{
					Name:  "override",
					Usage: "JSON file with key-value pairs for overriding non-secret values in Configstore DB",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
			BashComplete: ConfigstoreKeysAutocomplete,
		},
		{
			Name:   "ls",
			Usage:  "List all keys and their respective values from the Configstore",
			ArgsUsage: "[key_filter]",
			Action: cmdLs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.StringSliceFlag{
					Name:  "override",
					Usage: "JSON file with key-value pairs for overriding non-secret values in Configstore DB",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
				cli.BoolFlag{
					Name:  "skip-decryption",
					Usage: "Do not decrypt any secrets - these are replaced by \"(string)\"",
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
			BashComplete: ConfigstoreKeysAutocomplete,
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
				cli.StringSliceFlag{
					Name:  "override",
					Usage: "JSON file with key-value pairs for overriding non-secret values in Configstore DB",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
		{
			Name:      "test_template",
			Usage:     "Takes a GO template file, and checks that the provided Configstore has the required keys to fill it in",
			ArgsUsage: "/path/to/template",
			Action:    cmdTestTemplate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
			},
		},
		{
			Name:      "compare_keys",
			Usage:     "Takes two Configstore DB files, and checks that they both contain the same keys",
			ArgsUsage: "/path/to/database1 /path/to/database2",
			Action:    cmdCompareKeys,
		},
		{
			Name:      "exec",
			Usage:     "Execute a shell command which contains template variables",
			ArgsUsage: "command_template",
			Action:    cmdExec,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The Configstore JSON file",
					Value: "./configstore.json",
				},
				cli.StringSliceFlag{
					Name:  "override",
					Usage: "JSON file with key-value pairs for overriding non-secret values in Configstore DB",
				},
				cli.BoolFlag{
					Name:  "ignore-role",
					Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
				},
			},
		},
		{
			Name:    "package",
			Aliases: []string{"p"},
			Usage:   "Manage a Configstore package structure containing environments and templates",
			Subcommands: []cli.Command{
				{
					Name:      "init",
					Usage:     "Initialise Configstore package directory structure",
					ArgsUsage: "basedir",
					Action:    cmdPackageInit,
				},
				{
					Name:      "create_env",
					Usage:     "Create a Configstore DB for a new environment or sub-environment",
					ArgsUsage: "env[/subenv]",
					Action:    cmdPackageCreate,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
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
					Name:      "copy",
					Usage:     "Copy keys between two top-level environments",
					ArgsUsage: "source_env destination_env [key_pattern]",
					Action:    cmdPackageCopy,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "skip-decryption",
							Usage: "Do not decrypt any secrets - these are replaced by \"(string)\"",
						},
						cli.BoolFlag{
							Name:  "recursive",
							Usage: "Copy the selected environment, and all of its sub-environments recursively",
						},
						cli.BoolFlag{
							Name:  "skip-existing",
							Usage: "Do not overwrite values for existing keys in destination DB",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvNamesAutocomplete),
				},
				{
					Name:      "set",
					Usage:     "Set a new value, or update an existing one in the Configstore of a given environment",
					ArgsUsage: "env[/subenv] key value",
					Action:    cmdPackageSet,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "secret",
							Usage: "Whether this value is sensitive (to be encrypted)",
						},
						cli.BoolFlag{
							Name:  "binary",
							Usage: "Indicate whether this value contains binary data (instead of plain text)",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(nil),
				},
				{
					Name:      "get",
					Usage:     "Get a value from the Configstore in a given environment",
					ArgsUsage: "env[/subenv] key",
					Action:    cmdPackageGet,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvKeysAutocomplete),
				},
				{
					Name:      "unset",
					Usage:     "Remove a value from the Configstore for a given environment",
					ArgsUsage: "env[/subenv] key",
					Action:    cmdPackageUnset,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvKeysAutocomplete),
				},
				{
					Name:      "encrypt",
					Usage:     "Take an existing value from the given environment and change it to be a secret (encrypted) value",
					ArgsUsage: "env key",
					Action:    cmdPackageEncrypt,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvKeysAutocomplete),
				},
				{
					Name:      "decrypt",
					Usage:     "Take an existing value from the given environment and change it to be a plain (unencrypted) value",
					ArgsUsage: "env key",
					Action:    cmdPackageDecrypt,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvKeysAutocomplete),
				},
				{
					Name:      "ls",
					Usage:     "List key/value pairs from the given environment or sub-environment, or simply return a list of environments if no argument was given",
					ArgsUsage: "[env] [key_filter]",
					Action:    cmdPackageLs,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
						cli.BoolFlag{
							Name:  "skip-decryption",
							Usage: "Do not decrypt any secrets - these are replaced by \"(string)\"",
						},
					},
					BashComplete: PackageCmdAutocomplete(nil),
				},
				{
					Name:      "envs",
					Usage:     "Print out a hierarchical structure of all environments and sub-environments",
					Action:    cmdPackageEnvs,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
					},
					BashComplete: PackageCmdAutocomplete(nil),
				},
				{
					Name:   "diff",
					Usage:  "Prints out the differences between two environments",
					ArgsUsage: "env[/subenv] env[/subenv]",
					Action: cmdPackageDiff,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
						cli.BoolFlag{
							Name:  "skip-decryption",
							Usage: "Do not decrypt any secrets - these are replaced by \"(string)\"",
						},
					},
					BashComplete: PackageCmdAutocomplete(EnvNamesAutocomplete),
				},
				{
					Name:   "tree",
					Usage:  "Print out a hierarchical structure of all keys and values for environments and sub-environments",
					ArgsUsage: "[key_filter]",
					Action: cmdPackageTree,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
						cli.BoolFlag{
							Name:  "skip-decryption",
							Usage: "Do not decrypt any secrets - these are replaced by \"(string)\"",
						},
					},
				},
				{
					Name:      "process_templates",
					Usage:     "Process all template files using values from the environment and (optional) sub-environment provided",
					ArgsUsage: "env[/subenv] output_dir",
					Action:    cmdPackageProcessTemplates,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
						cli.BoolFlag{
							Name:  "ignore-role",
							Usage: "Do not assume the IAM Role for this Configstore (if one was set) before calling the AWS API",
						},
					},
					BashComplete: PackageCmdAutocomplete(nil),
				},
				{
					Name:   "test",
					Usage:  "Run checks on the Configstore DBs and templates in this package",
					Action: cmdPackageTest,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "basedir",
							Usage: "The base directory for the configuration package structure",
							Value: "./config",
						},
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
