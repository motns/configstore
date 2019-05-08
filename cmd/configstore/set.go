package main

import (
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"gopkg.in/urfave/cli.v1"
	"os"
	"io/ioutil"
	"github.com/motns/configstore/client"
)

func cmdSet(c *cli.Context) error {
	cc, err := client.NewConfigstoreClient(c.String("db"), "", c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	isSecret := c.Bool("secret")


	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Work out whether data is being piped in from StdIn

	var havePipe bool

	// Found the below two sections in a blog post here:
	//     https://coderwall.com/p/zyxyeg/golang-having-fun-with-os-stdin-and-shell-pipes
	ss, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if ss.Mode() & os.ModeNamedPipe != 0 {
		havePipe = true
	} else {
		havePipe = false
	}


	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Read raw value

	key := c.Args().Get(0)
	if key == "" {
		return errors.New("You have to specify a Key to set as the first argument")
	}

	var rawValue []byte

	if havePipe {
		rawValue, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		if isSecret {
			fmt.Print("Secret:")

			rawValue, err = gopass.GetPasswd()
			if err != nil {
				return err
			}

		} else {
			rawValue = []byte(c.Args().Get(1))
		}
	}


	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Process value if needed, then store

	err = cc.Set(key, rawValue, isSecret)
	if err != nil {
		return err
	}

	return nil
}
