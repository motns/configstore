package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/motns/configstore/client"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func SliceContains(s []string, el string) bool {
	for _, v := range s {
		if v == el {
			return true
		}
	}

	return false
}

func PrintLines(ls []string) {
	for _, s := range ls {
		println(s)
	}
}

func DirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ListDirs(basedir string) ([]string, error) {
	dirs := make([]string, 0)
	entries, err := ioutil.ReadDir(basedir)

	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	return dirs, nil
}

func ListFiles(basedir string) ([]string, error) {
	files := make([]string, 0)
	entries, err := ioutil.ReadDir(basedir)

	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		name := e.Name()
		e.Mode()
		if !e.IsDir() && name[0:1] != "." {
			files = append(files, name)
		}
	}

	return files, nil
}

func ParseEnv(s string, basedir string) (string, string, error)  {
	if s == "" {
		return "", "", errors.New("environment name cannot be empty")
	}

	var env string
	var subenv string

	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		env = parts[0]
		subenv = parts[1]
	} else {
		env = s
		subenv = ""
	}

	if basedir != "" {
		if EnvExists(basedir, env) == false {
			return "", "", errors.New("environment doesn't exists: " + env)
		}

		if subenv != "" {
			if SubEnvExists(basedir, env, subenv) == false {
				return "", "", errors.New("sub-environment doesn't exists: " + env + "/" + subenv)
			}
		}
	}

	return env, subenv, nil
}


func EnvExists(basedir string, env string) bool {
	return DirExists(basedir + "/env/" + env)
}


func SubEnvExists(basedir string, env string, subenv string) bool {
	return DirExists(basedir + "/env/" + env + "/subenv/" + subenv)
}


func LoadEnvOverride(basedir string, env string, subenv string) (map[string]string, error) {
	var overrides = make(map[string]string)

	jsonStr, err := ioutil.ReadFile(basedir + "/env/" + env + "/subenv/" + subenv + "/override.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonStr, &overrides); err != nil {
		return nil, err
	}

	return overrides, nil
}


func SaveEnvOverride(basedir string, env string, subenv string, override map[string]string) error {
	jsonStr, err := json.MarshalIndent(override, "", "  ")

	if err != nil {
		return err
	}

	return ioutil.WriteFile(basedir + "/env/" + env + "/subenv/" + subenv + "/override.json", jsonStr, 0644)
}


func ConfigstoreForEnv(basedir string, env string, subenv string, ignoreRole bool) (*client.ConfigstoreClient, error) {
	dbFile := basedir + "/env/" + env + "/configstore.json"

	overrideFiles := make([]string, 0)

	if subenv != "" {
		if SubEnvExists(basedir, env, subenv) == false {
			return nil, errors.New("sub-environment doesn't exists: " + env + "/" + subenv)
		}

		overrideFiles = append(overrideFiles, basedir + "/env/" + env + "/subenv/" + subenv + "/override.json")
	}

	cc, err := client.NewConfigstoreClient(dbFile, overrideFiles, ignoreRole)
	if err != nil {
		return nil, err
	}

	return cc, nil
}


func SetCmdShared(cc *client.ConfigstoreClient, isSecret bool, key string, val string) error {

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

	// Read raw value
	if key == "" {
		return errors.New("you have to specify a Key to set as the first argument")
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
			rawValue = []byte(val)
		}
	}

	err = cc.Set(key, rawValue, isSecret)
	if err != nil {
		return err
	}

	return nil
}


func RenderTable(allKeys []string, allValues map[string]map[string]string, envs []string, isSubEnv bool) {
	var headers []string

	if isSubEnv {
		headers = append([]string{"Key / SubEnv"}, envs...)
	} else {
		headers = append([]string{"Key / Env"}, envs...)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)

	for _, k := range allKeys {
		cols := make([]string, 0)
		cols = append(cols, k)

		firstVal := allValues[envs[0]][k]
		hasDiff := false

		for _, e := range envs {
			v := allValues[e][k]
			var formatted string

			if firstVal != v {
				hasDiff = true
			}

			if v == "" {
				formatted = formatRed("(missing)")
			} else {
				formatted = v
			}

			cols = append(cols, formatted)
		}

		if hasDiff {
			table.Append(formatAllYellow(cols))
		} else {
			table.Append(cols)
		}
	}

	table.Render()
}


func CompareKeys(cc1 *client.ConfigstoreClient, cc2 *client.ConfigstoreClient) ([]string, error) {
	out := make([]string, 0)

	db1Keys := cc1.GetAllKeys()
	db2Keys := cc2.GetAllKeys()

	notInDb1 := make([]string, 0)
	notInDb2 := make([]string, 0)

	for _, key := range db1Keys {
		if !SliceContains(db2Keys, key) {
			notInDb2 = append(notInDb2, key)
		}
	}

	for _, key := range db2Keys {
		if !SliceContains(db1Keys, key) {
			notInDb1 = append(notInDb1, key)
		}
	}

	if len(notInDb1) == 0 && len(notInDb2) == 0 {
		return nil, nil
	} else {
		if len(notInDb1) > 0 {
			out = append(out, "Keys not in DB 1:")
			sort.Strings(notInDb1)

			for _, key := range notInDb1 {
				out = append(out, "\""+key+"\"")
			}
		}

		if len(notInDb2) > 0 {
			out = append(out, "Keys not in DB 2:")
			sort.Strings(notInDb2)

			for _, key := range notInDb2 {
				out = append(out, "\""+key+"\"")
			}
		}

		return out, errors.New("databases did not match")
	}
}
