package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"sort"
)

func cmdLs(c *cli.Context) error {
	dbFile := c.String("db")
	db, err := loadConfigstore(dbFile)
	if err != nil {
		return err
	}

	enc, err := createEncryption(&db, nil, c.Bool("ignore-role"))
	if err != nil {
		return err
	}

	entries := make([]string, 0, len(db.Data))

	for key, entry := range db.Data {
		if entry.IsSecret {
			decrypted, err := enc.decrypt(entry.Value)
			if err != nil {
				return err
			}

			entries = append(entries, key+": "+decrypted)
		} else {
			entries = append(entries, key+": "+entry.Value)
		}
	}

	sort.Strings(entries)

	for _, v := range entries {
		fmt.Println(v)
	}

	return nil
}
