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

@test "configstore ls filter" {
  run bin/darwin/amd64/configstore ls --db test_data/example_configstore.json name
  [ "$status" -eq 0 ]
  [ "${lines[0]}" = "lastname: Parker" ]
  [ "${lines[1]}" = "username: admin" ]
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

@test "configstore encrypt and decrypt" {
  rm -f test_data/configstore.json
  bin/darwin/amd64/configstore init --dir test_data --insecure
  run bin/darwin/amd64/configstore set --db test_data/configstore.json mykey myvalue
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore encrypt --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore get --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]
  [ "$output" = "myvalue" ]

  run bin/darwin/amd64/configstore decrypt --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore get --db test_data/configstore.json mykey
  [ "$status" -eq 0 ]
  [ "$output" = "myvalue" ]

  rm -f test_data/configstore.json
}

@test "configstore test_template" {
  run bin/darwin/amd64/configstore test_template --db test_data/example_configstore.json test_data/valid_template.txt
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore test_template --db test_data/example_configstore.json test_data/invalid_template.txt
  [ "$status" -eq 1 ]
}

@test "configstore package creation" {
  rm -rf test_data/package_test
  rm -rf test_data/out_test
  mkdir test_data/out_test

  run bin/darwin/amd64/configstore package init test_data/package_test
  [ "$status" -eq 0 ]
  [ -d "test_data/package_test" ]
  [ -d "test_data/package_test/env" ]
  [ -d "test_data/package_test/template" ]

  cp test_data/valid_template.txt test_data/package_test/template/valid_template.txt

  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test dev
  [ "$status" -eq 0 ]
  [ -d "test_data/package_test/env/dev" ]
  [ -e "test_data/package_test/env/dev/configstore.json" ]

  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local
  [ "$status" -eq 0 ]
  [ -d "test_data/package_test/env/dev/local" ]
  [ -e "test_data/package_test/env/dev/local/override.json" ]

  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local/foo
  [ "$status" -eq 0 ]
  [ -d "test_data/package_test/env/dev/local/foo" ]
  [ -e "test_data/package_test/env/dev/local/foo/override.json" ]

  rm  -rf test_data/package_test
  rm -rf test_data/out_test
}

@test "configstore package get, set, unset" {
  rm -rf test_data/package_test
  rm -rf test_data/out_test
  mkdir test_data/out_test

  run bin/darwin/amd64/configstore package init test_data/package_test
  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test dev
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local/foo

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username root
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev username
  [ "$status" -eq 0 ]
  [ "$output" = "root" ]

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev password supersecret
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev password
  [ "$status" -eq 0 ]
  [ "$output" = "supersecret" ]

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local username admin
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev/local username
  [ "$status" -eq 0 ]
  [ "$output" = "admin" ]

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local/foo username kevin
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev/local/foo username
  [ "$status" -eq 0 ]
  [ "$output" = "kevin" ]

  run bin/darwin/amd64/configstore package unset --basedir test_data/package_test dev/local username
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev/local username
  [ "$status" -eq 0 ]
  [ "$output" = "root" ]

  run bin/darwin/amd64/configstore package unset --basedir test_data/package_test dev username
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test dev username
  [ "$status" -eq 1 ]
  [ "$output" = "key does not exist in Configstore: username" ]

  rm  -rf test_data/package_test
  rm -rf test_data/out_test
}

@test "configstore package ls, envs, tree, diff" {
  rm -rf test_data/package_test
  rm -rf test_data/out_test
  mkdir test_data/out_test

  run bin/darwin/amd64/configstore package init test_data/package_test
  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test dev
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local/foo

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username root
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev password supersecret
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local username admin
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local/foo username kevin

  run bin/darwin/amd64/configstore package ls --basedir test_data/package_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package ls --basedir test_data/package_test dev
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package ls --basedir test_data/package_test dev/local
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package envs --basedir test_data/package_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package tree --basedir test_data/package_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package tree user --basedir test_data/package_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package diff --basedir test_data/package_test dev dev/local
  [ "$status" -eq 0 ]

  rm  -rf test_data/package_test
  rm -rf test_data/out_test
}

@test "configstore package test, process_template" {
  rm -rf test_data/package_test
  rm -rf test_data/out_test
  mkdir test_data/out_test

  run bin/darwin/amd64/configstore package init test_data/package_test
  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test dev
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local/foo

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username root
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev password supersecret
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local username admin
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local/foo username kevin

  run bin/darwin/amd64/configstore package test --basedir test_data/package_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package process_templates --basedir test_data/package_test dev test_data/out_test
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package process_templates --basedir test_data/package_test dev/local test_data/out_test
  [ "$status" -eq 0 ]

  rm  -rf test_data/package_test
  rm -rf test_data/out_test
}

@test "configstore package copy" {
  rm -rf test_data/package_test
  rm -rf test_data/out_test
  mkdir test_data/out_test

  run bin/darwin/amd64/configstore package init test_data/package_test
  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test dev
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local
  run bin/darwin/amd64/configstore package create_env --basedir test_data/package_test dev/local/foo

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username root
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev password supersecret
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local username admin
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev/local/foo username kevin

  run bin/darwin/amd64/configstore package create_env --insecure --basedir test_data/package_test staging

  run bin/darwin/amd64/configstore package copy --basedir test_data/package_test --recursive dev staging
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging username
  [ "$status" -eq 0 ]
  [ "$output" = "root" ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging password
  [ "$status" -eq 0 ]
  [ "$output" = "supersecret" ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging/local username
  [ "$status" -eq 0 ]
  [ "$output" = "admin" ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging/local/foo username
  [ "$status" -eq 0 ]
  [ "$output" = "kevin" ]

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username dave

  run bin/darwin/amd64/configstore package copy --basedir test_data/package_test dev staging
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging username
  [ "$status" -eq 0 ]
  [ "$output" = "dave" ]

  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev username john
  run bin/darwin/amd64/configstore package set --basedir test_data/package_test dev url "https://www.example.com"

  run bin/darwin/amd64/configstore package copy --basedir test_data/package_test --skip-existing dev staging
  [ "$status" -eq 0 ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging username
  [ "$status" -eq 0 ]
  [ "$output" = "dave" ]

  run bin/darwin/amd64/configstore package get --basedir test_data/package_test staging url
  [ "$status" -eq 0 ]
  [ "$output" = "https://www.example.com" ]

  rm  -rf test_data/package_test
  rm -rf test_data/out_test
}