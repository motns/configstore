package main

import (
	"github.com/motns/configstore/client"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)

func cmdPackageDiff(c *cli.Context) error {
	basedir := c.String("basedir")
	ignoreRole := c.Bool("ignore-role")
	skipDecryption := c.Bool("skip-decryption")
	env1Str := c.Args().Get(0)
	env2Str := c.Args().Get(1)

	env1, err := ParseEnv(env1Str, basedir, true)
	if err != nil {
		return err
	}

	env2, err := ParseEnv(env2Str, basedir, true)
	if err != nil {
		return err
	}

	cc1, err := ConfigstoreForEnv(env1, ignoreRole)
	if err != nil {
		return err
	}

	cc2, err := ConfigstoreForEnv(env2, ignoreRole)
	if err != nil {
		return err
	}

	cc1Map, err := cc1.GetAll(skipDecryption)
	if err != nil {
		return err
	}

	cc2Map, err := cc2.GetAll(skipDecryption)
	if err != nil {
		return err
	}

	keySet := make(map[string]int)

	for k := range cc1Map {
		keySet[k] = 1
	}

	for k := range cc2Map {
		keySet[k] = 1
	}

	allKeys := make([]string, 0)

	for k := range keySet {
		allKeys = append(allKeys, k)
	}

	sort.Strings(allKeys)

	renderDiffTable([]string{"Key", env1Str, env2Str}, allKeys, cc1Map, cc2Map)

	return nil
}

func renderDiffTable(header []string, allKeys []string, cc1Map map[string]client.ConfigstoreDBValue, cc2Map map[string]client.ConfigstoreDBValue) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	hasDiff := false

	for _, k := range allKeys {
		str1 := getFormattedValue(k, cc1Map)
		str2 := getFormattedValue(k, cc2Map)

		if str1 != str2 {
			hasDiff = true
			table.Append([]string{k, str1, str2})
		}
	}

	if hasDiff {
		table.Render()
	} else {
		println("The two DBs match")
	}
}

func getFormattedValue(key string, ccMap map[string]client.ConfigstoreDBValue) string {
	val, exists := ccMap[key]

	if !exists {
		return formatRed("(missing)")
	} else {
		var str = ""

		if val.IsBinary {
			str = "(binary)"
		} else {
			str = val.Value
		}

		if val.IsSecret {
			return formatYellow(str)
		} else {
			return str
		}
	}
}