package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"github.com/motns/configstore/client"
	"errors"
	"sort"
)

func cmdCompareKeys(c *cli.Context) error {
	dbFile1 := c.Args().Get(0)
	if dbFile1 == "" {
		return errors.New("you have to provide a Configstore DB as the first argument")
	}

	dbFile2 := c.Args().Get(1)
	if dbFile2 == "" {
		return errors.New("you have to provide a Configstore DB to compare against as the second argument")
	}

	cc1, err := client.NewConfigstoreClient(dbFile1, "", c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	cc2, err := client.NewConfigstoreClient(dbFile2, "", c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	db1Keys := cc1.GetAllKeys()
	db2Keys := cc2.GetAllKeys()

	notInDb1 := make([]string, 0)
	notInDb2 := make([]string, 0)

	for _, key := range db1Keys {
		if !contains(db2Keys, key) {
			notInDb2 = append(notInDb2, key)
		}
	}

	for _, key := range db2Keys {
		if !contains(db1Keys, key) {
			notInDb1 = append(notInDb1, key)
		}
	}

	if len(notInDb1) == 0 && len(notInDb2) == 0 {
		fmt.Println("Databases match")
		return nil
	} else {
		if len(notInDb1) > 0 {
			fmt.Println("Keys not in DB 1:")
			sort.Strings(notInDb1)

			for _, key := range notInDb1 {
				fmt.Println("\""+key+"\"")
			}
		}

		if len(notInDb2) > 0 {
			fmt.Println("Keys not in DB 2:")
			sort.Strings(notInDb2)

			for _, key := range notInDb2 {
				fmt.Println("\""+key+"\"")
			}
		}

		return errors.New("databases did not match")
	}
}

func contains(s []string, el string) bool {
	for _, v := range s {
		if v == el {
			return true
		}
	}

	return false
}