#!/usr/bin/env bats

@test "configstore compare-keys" {
  run bin/darwin/amd64/configstore compare_keys test_data/example_configstore.json test_data/example_configstore.json
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore compare_keys test_data/example_configstore.json test_data/example_configstore_two.json
  [ "$status" -eq 1 ]
  [ "${lines[0]}" = "Keys not in DB 1:" ]
  [ "${lines[1]}" = "\"firstname\"" ]
  [ "${lines[2]}" = "Keys not in DB 2:" ]
  [ "${lines[3]}" = "\"lastname\"" ]
  [ "${lines[4]}" = "databases did not match" ]
}

@test "configstore get" {
  run bin/darwin/amd64/configstore get --db test_data/example_configstore.json username
  [ "$status" -eq 0 ]
  [ "$output" = "admin" ]
}

@test "configstore get override" {
  run bin/darwin/amd64/configstore get --db test_data/example_configstore.json --override test_data/override.json email
  [ "$status" -eq 0 ]
  [ "$output" = "peter.parker@example.com" ]
}

@test "configstore init" {
  rm -f test_data/configstore.json
  run bin/darwin/amd64/configstore init --dir test_data --insecure
  [ "$status" -eq 0 ]
  rm -f test_data/configstore.json
}

@test "configstore ls" {
  run bin/darwin/amd64/configstore ls --db test_data/example_configstore.json
  [ "$status" -eq 0 ]
  [ "${lines[0]}" = "email: spider-man@example.com" ]
  [ "${lines[1]}" = "lastname: Parker" ]
  [ "${lines[2]}" = "password: supersecret" ]
  [ "${lines[3]}" = "username: admin" ]
}

@test "configstore process_template" {
  run bin/darwin/amd64/configstore process_template --db test_data/example_configstore.json test_data/valid_template.txt
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore process_template --db test_data/example_configstore.json test_data/invalid_template.txt
  [ "$status" -eq 1 ]
}

@test "configstore set and unset" {
  rm -f test_data/configstore.json
  bin/darwin/amd64/configstore init --dir test_data --insecure
  run bin/darwin/amd64/configstore set --db test_data/configstore.json mykey myvalue
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore get --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]
  [ "$output" = "myvalue" ]

  run bin/darwin/amd64/configstore unset --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore get --db test_data/configstore.json mykey
  [ "$status" -eq 1 ]
  [ "$output" = "key does not exist in Configstore: mykey" ]

  rm -f test_data/configstore.json
}

@test "configstore test_template" {
  run bin/darwin/amd64/configstore test_template --db test_data/example_configstore.json test_data/valid_template.txt
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore test_template --db test_data/example_configstore.json test_data/invalid_template.txt
  [ "$status" -eq 1 ]
}