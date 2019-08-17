package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/motns/configstore/client"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)


///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Data Structure for storing the path to an environment

type Env struct {
	basedir     string
	envName     string
	subenvNames []string
}

func (e *Env) mainEnvPath() string {
	return e.basedir + "/env/" + e.envName
}

func (e *Env) envPath() string {
	var out = e.basedir + "/env/" + e.envName

	for _, s := range e.subenvNames {
		out = out + "/" + s
	}

	return out
}

func (e *Env) envStr() string {
	var out = e.envName

	for _, s := range e.subenvNames {
		out = out + "/" + s
	}

	return out
}

func (e *Env) mainEnvExists() bool {
	return DirExists(e.mainEnvPath())
}

func (e *Env) envExists() bool {
	return DirExists(e.envPath())
}

func (e *Env) isSubenv() bool {
	return len(e.subenvNames) > 0
}

func (e *Env) isMainEnv() bool {
	return len(e.subenvNames) == 0
}

func (e *Env) dbFile() string {
	return e.mainEnvPath() + "/configstore.json"
}

func (e *Env) overrideFile() (string, error) {
	if !e.isSubenv() {
		return "", errors.New("trying to get override file for top-level environment")
	}

	return e.envPath() + "/override.json", nil
}

func (e *Env) overrideFiles() []string {
	overrideFiles := make([]string, 0)

	for k := range e.subenvNames {
		overrideFiles = append(overrideFiles, e.basedir + "/env/" + e.envName+ "/" + strings.Join(e.subenvNames[0:k+1], "/") + "/override.json")
	}

	return overrideFiles
}

func (e *Env) getSubenv(subenv string) Env {
	return Env{
		basedir: e.basedir,
		envName: e.envName,
		subenvNames: append(e.subenvNames, subenv),
	}
}

func ParseEnv(s string, basedir string, validate bool) (Env, error) {
	if s == "" {
		return Env{}, errors.New("environment name cannot be empty")
	}

	var env string
	var subenvs []string

	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		env = parts[0]
		subenvs = parts[1:]
	} else {
		env = s
	}

	e := Env{
		basedir:     basedir,
		envName:     env,
		subenvNames: subenvs,
	}

	if validate && !e.envExists() {
		return Env{}, errors.New("environment does not exist: " + e.envStr())
	}

	return e, nil
}


///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers

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

func LoadEnvOverride(basedir string) (map[string]string, error) {
	var overrides = make(map[string]string)

	jsonStr, err := ioutil.ReadFile(basedir + "/override.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonStr, &overrides); err != nil {
		return nil, err
	}

	return overrides, nil
}

func SaveEnvOverride(basedir string, override map[string]string) error {
	jsonStr, err := json.MarshalIndent(override, "", "  ")

	if err != nil {
		return err
	}

	return ioutil.WriteFile(basedir+"/override.json", jsonStr, 0644)
}

func ConfigstoreForEnv(env Env, ignoreRole bool) (*client.ConfigstoreClient, error) {
	cc, err := client.NewConfigstoreClient(env.dbFile(), env.overrideFiles(), ignoreRole)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func SetCmdShared(cc *client.ConfigstoreClient, isSecret bool, isBinary bool, key string, val string) error {

	// Work out whether data is being piped in from StdIn
	var havePipe bool

	// Found the below two sections in a blog post here:
	//     https://coderwall.com/p/zyxyeg/golang-having-fun-with-os-stdin-and-shell-pipes
	ss, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if ss.Mode()&os.ModeNamedPipe != 0 {
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

	err = cc.Set(key, rawValue, isSecret, isBinary)
	if err != nil {
		return err
	}

	return nil
}

func CreateSubenvShared(env Env) error {
	if !env.mainEnvExists() {
		return errors.New("main environment doesn't exist: " + env.envName)
	}

	if env.envExists() {
		return errors.New("sub-environment already exists: " + env.envStr())
	}

	if err := os.MkdirAll(env.envPath(), 0755); err != nil {
		return err
	}

	filePath, err := env.overrideFile()

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	if _, err := f.WriteString("{}"); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	return nil
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
