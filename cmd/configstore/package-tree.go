package main

import (
	"errors"
	"github.com/motns/configstore/client"
	"gopkg.in/urfave/cli.v1"
	"sort"
	"strings"
)

type Tree = map[string]TreeNode

type TreeNode struct {
	value    string
	children Tree
}

func cmdPackageTree(c *cli.Context) error {
	basedir := c.String("basedir")
	ignoreRole := c.Bool("ignore-role")
	skipDecryption := c.Bool("skip-decryption")

	envs, err := ListDirs(basedir + "/env")
	if err != nil {
		return err
	}

	configstores, err := loadAllConfigstores(basedir, envs, ignoreRole)
	if err != nil {
		return err
	}

	allKeys := getAllConfigstoreKeys(configstores)
	configTree, err := buildTree(basedir, configstores, allKeys, skipDecryption)
	if err != nil {
		return err
	}

	printTree(configTree, 0, true)

	return nil
}

func buildTree(basedir string, configstores map[string]*client.ConfigstoreClient, allKeys []string, skipDecryption bool) (Tree, error) {
	configTree := make(Tree)
	cache := createCache()

	for env, cc := range configstores {
		entries, err := cc.GetAll(skipDecryption)

		if err != nil {
			return nil, err
		}

		for _, k := range allKeys {

			if _, exists := configTree[k]; !exists {
				configTree[k] = TreeNode{
					value:    "",
					children: make(map[string]TreeNode),
				}
			}

			var val string
			entry, exists := entries[k]

			if exists {
				if entry.IsBinary {
					val = "(binary)"
				} else {
					val = entry.Value
				}

				if entry.IsSecret {
					val = formatYellow(val)
				}
			} else {
				val = formatRed("(missing)")
			}

			subtree, hasOverride, err := buildSubTree(k, basedir+"/env/"+env, cache)
			if err != nil {
				return nil, err
			}

			// Blank out subtree if it contains to overrides for this key, to reduce noise in output
			if !hasOverride {
				subtree = nil
			}

			configTree[k].children[env] = TreeNode{
				value:    val,
				children: subtree,
			}
		}
	}

	return configTree, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Process Configstores

func loadAllConfigstores(basedir string, envs []string, ignoreRole bool) (map[string]*client.ConfigstoreClient, error) {
	configstores := make(map[string]*client.ConfigstoreClient)

	for _, env := range envs {
		cc, err := ConfigstoreForEnv(basedir, env, nil, ignoreRole)

		if err != nil {
			return nil, errors.New("failed to load Configstore \"" + env + "\": " + err.Error())
		}

		configstores[env] = cc
	}

	return configstores, nil
}

func getAllConfigstoreKeys(configstores map[string]*client.ConfigstoreClient) []string {
	keySet := make(map[string]int)

	for _, cc := range configstores {
		keys := cc.GetAllKeys()

		for _, k := range keys {
			keySet[k] = 1
		}
	}

	allKeys := make([]string, 0)
	for k := range keySet {
		allKeys = append(allKeys, k)
	}

	return allKeys
}

func buildSubTree(key string, basedir string, cache SubenvCache) (Tree, bool, error) {
	subenvs, err := ListDirs(basedir)
	if err != nil {
		return nil, false, err
	}

	if len(subenvs) == 0 {
		return nil, false, nil
	}

	tree := make(Tree)
	var treeHasOverride = false
	var subtreeHasOverride = false

	for _, subenv := range subenvs {
		data, err := cache.get(basedir + "/" + subenv)
		if err != nil {
			return nil, false, err
		}

		var val string
		val, exists := data[key]

		if !exists {
			val = ""
		} else {
			treeHasOverride = true
		}

		subtree, o, err := buildSubTree(key, basedir+"/"+subenv, cache)
		subtreeHasOverride = subtreeHasOverride || o
		if err != nil {
			return nil, false, err
		}

		// Drop subtree if it doesn't contain any overrides for this key, to remove noise in the output
		if !o {
			subtree = nil
		}

		tree[subenv] = TreeNode{
			value:    val,
			children: subtree,
		}
	}

	return tree, treeHasOverride || subtreeHasOverride, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Subenv cache

type SubenvCache struct {
	cache map[string]map[string]string
}

func createCache() SubenvCache {
	return SubenvCache{
		cache: make(map[string]map[string]string),
	}
}

func (c SubenvCache) get(basedir string) (map[string]string, error) {
	var data map[string]string
	data, exists := c.cache[basedir]

	if !exists {
		override, err := LoadEnvOverride(basedir)
		if err != nil {
			return nil, err
		}
		c.cache[basedir] = override
		data = override
	}

	return data, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Output

func printTree(configTree Tree, indent int, isRoot bool) {
	allKeys := make([]string, 0)

	for key := range configTree {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		printTreeNode(key, configTree[key], indent, isRoot)
	}
}

func printTreeNode(key string, node TreeNode, indent int, isRoot bool) {
	out := strings.Repeat(" ", indent)
	if !isRoot {
		out += formatCyan("/" + key)
	} else {
		out += formatGreen(key)
	}

	if node.value != "" {
		out += ": " + node.value
	}

	println(out)

	if len(node.children) != 0 {
		printTree(node.children, indent+2, false)
	}
}
