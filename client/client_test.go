package client

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestInitConfigstore(t *testing.T) {
	// Remove previous test data file (ignore errors)
	os.Remove("../test_data/configstore.json")

	// Test successful initialisation first, via insecure mode
	_, err := InitConfigstore("../test_data", "eu-west-1", "", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	jsonStr, err := ioutil.ReadFile("../test_data/configstore.json")

	if err != nil {
		t.Errorf("could not read newly created configstore file: %s", err)
	}

	if len(jsonStr) == 0 {
		// TODO - Validate actual contents
		t.Error("newly created configstore file is empty")
	}

	// Test error case, with missing master key
	_, err = InitConfigstore("../test_data", "eu-west-1", "", "", false)

	if err == nil {
		t.Error("expected configstore init to fail due to missing key")
	}

	// Clean up after ourselves
	os.Remove("../test_data/configstore.json")
}

func TestNewConfigstoreClient(t *testing.T) {
	_, err := NewConfigstoreClient("../test_data/example_configstore.json", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}
}

func TestGet(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	username, err := c.Get("username")

	if err != nil {
		t.Errorf("failed to get username key: %s", err)
	}

	if username != "admin" {
		t.Errorf("expected \"admin\" got %s", username)
	}

	password, err := c.Get("password")

	if err != nil {
		t.Errorf("failed to get password key: %s", err)
	}

	if password != "supersecret" {
		t.Errorf("expected \"supersecret\" got %s", password)
	}

	_, err = c.Get("foo")

	if err == nil {
		t.Error("expected get for missing key to fail")
	}
}

func TestProcessTemplateString(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	out, err := c.ProcessTemplateString("some_command {{.username}} {{.password}}")

	if err != nil {
		t.Errorf("failed to process template: %s", err)
	}

	if out != "some_command admin supersecret" {
		t.Errorf("expected \"some_command admin supersecret\" got \"%s\"", out)
	}
}

func TestTestTemplateString(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	result1, err := c.TestTemplateString("some_command {{.username}} {{.password}}")

	if err != nil {
		t.Errorf("failed to test template: %s", err)
	}

	if result1 != true {
		t.Error("expected true got false")
	}

	result2, err := c.TestTemplateString("some_command {{.foo}}")

	if err == nil {
		t.Error("expected \"template: tmp:1:15: executing \"tmp\" at <.foo>: map has no entry for key \"foo\"\" got nil")
	}

	if result2 != false {
		t.Error("expected false got true")
	}
}

func TestGetWithOverride(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "../test_data/override.json", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	username, err := c.Get("username")

	if err != nil {
		t.Errorf("failed to get username key: %s", err)
	}

	if username != "root" {
		t.Errorf("expected \"root\" got %s", username)
	}
}

func TestGetAll(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	configMap, err := c.GetAll()

	if err != nil {
		t.Errorf("failed to load configstore values: %s", err)
	}

	if len(configMap) != 4 {
		t.Errorf("expected 4 elements in configmap, got %d", len(configMap))
	}

	if configMap["lastname"] != "Parker" {
		t.Errorf("expected \"Parker\" got %s", configMap["lastname"])
	}

	if configMap["username"] != "admin" {
		t.Errorf("expected \"admin\" got %s", configMap["username"])
	}

	if configMap["password"] != "supersecret" {
		t.Errorf("expected \"supersecret\" got %s", configMap["password"])
	}

	if configMap["email"] != "spider-man@example.com" {
		t.Errorf("expected \"spider-man@example.com\" got %s", configMap["email"])
	}
}

func TestGetAllWithOverride(t *testing.T) {
	c, err := NewConfigstoreClient("../test_data/example_configstore.json", "../test_data/override.json", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	configMap, err := c.GetAll()

	if err != nil {
		t.Errorf("failed to load configstore values: %s", err)
	}

	if len(configMap) != 4 {
		t.Errorf("expected 4 elements in configmap, got %d", len(configMap))
	}

	if configMap["email"] != "peter.parker@example.com" {
		t.Errorf("expected \"peter.parker@example.com\" got %s", configMap["email"])
	}
}

// More of an integration test, exercising all (remaining) methods
func TestClientFull(t *testing.T) {
	// Remove previous test data file (ignore errors)
	os.Remove("../test_data/configstore.json")

	// Test successful initialisation first, via insecure mode
	c, err := InitConfigstore("../test_data", "eu-west-1", "", "", true)

	if err != nil {
		t.Errorf("failed to initialise configstore client: %s", err)
	}

	err = c.Set("realname", []byte("Peter Parker"), true)

	if err != nil {
		t.Errorf("failed to set new key: %s", err)
	}

	v, err := c.Get("realname")

	if err != nil {
		t.Errorf("failed to read value: %s", err)
	}

	if v != "Peter Parker" {
		t.Errorf("expected \"Peter Parker\" got %s", v)
	}

	err = c.Unset("realname")

	if err != nil {
		t.Errorf("failed to unset key: %s", err)
	}

	_, err = c.Get("realname")

	if err.Error() != "key does not exist in Configstore: realname" {
		t.Error("expected for key Get to fail but it did not")
	}

	// Clean up after ourselves
	os.Remove("../test_data/configstore.json")
}