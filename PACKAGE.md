# Configstore Package

A "Configstore package" is just a practical implementation of Configstore, where a number of "environments" (Configstore DBs),
and template files are collected in a specific directory structure. You may choose to make use of Configstore in a different
way in your application, but if you go with this "package" format, then you are able to make use of a number of built-in
commands for managing things.

The majority of the functionality under the `configstore package` command is built on top of existing commands and client
methods, so you could choose to implement your own version of managing DBs and templates, if you don't agree with how
we've done it here.

A Configstore Package organises Configstore DBs into "environments" (one env == one DB), with zero or more "sub-environments" under
each, which are basically just override files for the DB in that particular environment (you can read more about Overrides in
the general usage documentation [here](USAGE.md#Overrides)). You then have one or more template files, which are processed
in the context of one of these environments when you run Configstore in your application.

**Note** that one of the rules for a Configstore Package is that all environment Configstore DBs have to contain **exactly
the same keys** (but not the same values). This is so that package templates can be processed in any of the environments. 

The directory structure looks something like this:
```
/package
  /env
    /dev
      configstore.json
      /subenv
        /local
          override.json
        /aws
          override.json
    /staging
      configstore.json
    /live
      configstore.json
  /template
    application.conf
``` 


### Usage

Many of the commands are based on what's available for managing a regular Configstore DB (`get`, `set`, `ls`, etc.); the
`package` command just simplifies the process of executing them in the context of a given Configstore DB.

To initialise a new Configstore Package, just run:
```bash
configstore package init /path/to/package
```
This will create a skeleton directory structure at the given path.

To create an environment (with an insecure DB), run:
```bash
configstore package create_env dev --insecure
```
Note how we pass the name of the environment as the first argument - this will create a Configstore at `package/env/dev`.
`create_env` behaves pretty much exactly the same way as `configstore init`, except that it doesn't take a `--dir` argument.

Creating a sub-environment looks very similar:
```bash
configstore package create_env dev/local
```
This will create an empty override file under `package/env/dev/subenv/local`.
Note how you refer to sub-environments in a path like syntax (`env/subenv`); this is the same in all `configstore package` commands.

To set a new key in an environment:
```bash
configstore package set dev username admin
```

And to set an override for a sub-environment:
```bash
configstore package set dev/local username kevin
```
Note that trying to set keys in an uninitialised environment or sub-environment will result in an error.

There's also a fancier version of `ls` available, which renders environments in a table format:
```bash
configstore package ls
```
This will gather keys/values from all environments (but not sub-environments), and render them in a table that looks something like this:
```bash
+-----------+---------------------------+-----------------------------+-------------------------------+
| KEY / ENV |            DEV            |            LIVE             |            STAGING            |
+-----------+---------------------------+-----------------------------+-------------------------------+
| password  | password123               | supersecret                 | bases7-prank                  |
| url       | http://dev.myserver.org   | http://www.myserver.co.uk   | http://staging.myserver.org   |
| username  | admin                     | admin                       | admin                         |
+-----------+---------------------------+-----------------------------+-------------------------------+
```
It highlights entries which are missing from one or more environment, and also rows where the values are different between
environments.

You can run the same command for a sub-environment by passing the environment name as the first argument:
```bash
configstore package ls dev
```
This will load the environment, along with all sub-environments, and output a table similar to the above:
```bash
+--------------+---------------------------+-------------+-----------+
| KEY / SUBENV |         (DEFAULT)         |     AWS     |   LOCAL   |
+--------------+---------------------------+-------------+-----------+
| password     | password123               | supersecret | (missing) |
| url          | http://dev.myserver.org   | (missing)   | (missing) |
| username     | admin                     | (missing)   | kevin     |
+--------------+---------------------------+-------------+-----------+
```

To process templates in the context of a given environment, you can run:
```bash
configstore package process_templates dev/local /path/to/output
```
This will load the Configstore DB from `package/env/dev` with the override file from `package/env/dev/subenv/local`, and then process each
template file under `package/template`, outputting the result (with the same filenames) under `/path/to/output`.

To "test" a Configstore Package, you can run:
```bash
configstore package test
```
This will check that:
1. All environment Configstore DBs contain the exact same set of keys
2. Each template file is valid, and only contains keys referenced in the Configstore DBs

This command is implemented in a way that is doesn't need to decrypt the actual values from each Configstore, which means
that you can run it on a CI server as part of your build.


### Integrating with your application (using Docker)

It's entirely up to you how you want to make use of this feature, but a recommended way is to have a Configstore Package
structure as part of your application source (in a separate folder, maybe `devops/config`). You then copy this folder
into the Docker image when building your app; this allows you to use the same built Docker image in each environment.
All you need in your Docker image is a baked-in version of `configstore`, and an environment variable (maybe `$CONFIG_ENV`) to specify which
environment you're configuring. Then, as part of your Docker entry point, you create a script which calls
`configstore package process_templates $CONFIG_ENV /etc/blurb` first, and then calls your actual entry point. Your application
can then read the config files from inside `/etc/blurb`.

Additionally, you should include `configstore package test` in your CI pipeline to reduce the likelihood that Configstore
is unable to run `process_templates` when it comes to starting the container. This of course doesn't eliminate potential
permission issues with your KMS key on your chosen server, but there's currently no way to pre-verify that.
