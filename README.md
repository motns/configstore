# Configstore

A simple command line application written in Go, used to store a mixture of plain-text and encrypted values, in a plain-text
JSON file. This JSON file is safe to commit to version control.

The encryption uses [AWS KMS](https://aws.amazon.com/kms/) as a starting point: A generated **Data Key** is used to
encrypt the secrets themselves (AES 256), which is then itself encrypted using a KMS **Master Key** and stored in the
JSON file alongside the values.
Whenever the plain text value of a secret needs to be loaded, the **Data Key** is decrypted via AWS KMS, and that key
is then used to decrypt the secret value. The decrypted **Data Key** is then discarded; it is never stored in plain form.

### Initial Setup

In order to use the config store, it first needs to be initialised. However, before you can do that, you need to create
an AWS KMS Master Key - [see Documentation](http://docs.aws.amazon.com/kms/latest/developerguide/create-keys.html).

You'll also need to grab a copy of the `configstore` binary for your platform from [Releases](https://github.com/CultBeauty/configstore/releases),
and place it somewhere on your `$PATH` (perhaps `/usr/local/bin`). 
The project is currently compiled for 64bit Mac OS, Linux and Windows.

Finally, you need to have AWS API credentials configured; since Configstore uses the AWS Go SDK for making requests, it will support all the usual methods (environment variables, credentials file, EC2 role, etc.). You can read more about these on the [SDK Configuration page](http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).
If you're unsure what any of the above means, the quickest way to get started is to create a `~/.aws/credentials` file with the following contents (filling in the API credentials you got from the AWS Console):
```
[default]
aws_access_key_id = YOURACCESSKEY
aws_secret_access_key = YourSecretKey
```
You may already have this file if you installed the AWS CLI tool, in which case there's nothing else to do.

When you have these ready, you can run `configstore init`, specifying the Master Key ARN or Alias
(the Key needs to be in the specified region). For example:
```bash
configstore init --master-key "alias/my-key"
```
The AWS Region defaults to `eu-west-1`; you may specify a different Region via the `--region` option.

The above will generate and immediately encrypt a new **Data Key** using the KMS **Master Key**
you specified, and then create a `configstore.json` that looks something like this:
```json
{
  "version": 1,
  "region": "eu-west-1",
  "data_key":"your encrypted data key here",
  "data": {}
}
```
The `configstore.json` is created in the current directory by default, but you may also specify a different
path via the `--dir` option.


### Storing and Retrieving Values

To store a new plain text value:
```bash
configstore set <key> <value>
```
for example:
```bash
configstore set username root
```

To store a secret value (it will be encrypted before storage):
```bash
configstore set --secret <key>
```
Instead of taking the value as an argument, the app will prompt you to type it in with a silent input for security.

> WARNING: The silent input method above doesn't currently support multi-byte characters!

You may also pipe data into the app, instead of passing the value as an argument. For example:
```bash
cat myfile.txt | configstore set myfile
```
This also works for storing secrets, allowing you to pipe the data from a file, instead of pasting it into a prompt:
```bash
cat supersecret.txt | configstore set --secret secretfile
```
> WARNING: You should never use this facility in combination with `echo` or similar
(like `echo "mypassword" | configstore set --secret mypass`), since it potentially exposes your secret to other
users on the system!

To retrieve a value:
```bash
configstore get <key>
```
This will print out plain text config items as is, whereas secrets are decrypted first,
and then printed out.

To list all keys with their respective values:
```bash
configstore ls
```
Like `get`, this will decrypt secret values before printing them out.

To remove a value:
```bash
configstore unset <key>
```
This will delete the given item from the `configstore.json` file.

You can take two Configstore databases, and make sure that they both contain the exact same set of keys by calling:
```bash
configstore compare_keys one/configstore.json two/configstore.json
```
This will return exit code `0` if they both have the same keys, or exit code `1` if they don't, along with a list of keys missing from
the first or the second DB. Please note that this command only compares the list of keys, and **not** the actual values
associated with those keys.
This command is useful in cases where you have multiple Configstore DBs (one for each environment for example), and you
need to keep the keys lined up. Also, since it doesn't need to read the underlying values (and therefore decrypt via AWS KMS),
it can be run without special permissions, or even AWS access, which can be convenient on a CI server.

Finally, you can retrieve a key and encrypt it using the KMS Master key associated with the given Configstore by calling:
```bash
configstore as_kms_enc <key>
```
This is useful for example if you want to use one of your secret values with AWS Lambda. Lambda has support for taking
encrypted secrets via its configuration, which it then decrypts and passes to your function as an environment variable
at run time. Unfortunately it only support KMS master keys when doing this, whereas Configstore normally uses a generated
data key internally.


### Using template files

The app also has the ability to take a template file, and fill in values from the Configstore DB. The app supports Go's
plain-text template format - you can read more about it [here](https://golang.org/pkg/text/template/).

A simple example would be to create a template file `application.conf` with contents:
```
settingA = {{.foo}}
settingB = {{.bar}}
```
and then call the Configstore app with:
```bash
configstore process_template application.conf
```
The app then loads all values from the Configstore DB (decrypting secret values in the process), and then attempts
to process the template file, replacing `{{.foo}}` and `{{.bar}}` in our example with the values stored under those keys (`foo`, `bar`).
If you reference a key in your template file that doesn't exist in the Configstore, the app will raise an error.
The processed output is then sent to `stdout`, so you can just pipe it into another file for storage. 

You also have the option of just "testing" a template, which checks that every key referenced in the template is
available in the Configstore. You do this by calling:
```bash
configstore test_template application.conf
``` 
This command doesn't actually output the rendered template; it returns exit code `0` if the test was successful,
or exit code `1`, along with an error message, if the test failed. It's also worth noting that the app doesn't
actually try to decode the values in this mode; it simply grabs all the available keys from the Configstore,
and feeds them through to the template engine with dummy values. This is useful, since it allows you to run the test
anywhere (for example on a CI server), even if you don't have permissions to use the AWS KMS key for decrypting values.


### Overrides

Overrides are helpful in cases where you have a single Configstore for an environment, but you need two or more versions
of it with only minor differences. Instead of having to duplicate the entire Configstore DB, you can override one or more
keys via an override file, which is basically just a JSON file with a single, flat object in it, where both keys and values
are strings. 

Overrides are supported for any command which reads or outputs data (`get`, `ls`, `process_template`, etc.) via the
`--override` flag.

The rules for overrides are:
 * You can only apply one override file at a time
 * Overrides cannot contain any encrypted values, and therefore can't be used to override secrets. If you have different secret values,
   you should have different Configstore DBs instead.
 * You can't use the override to append new keys, and your override can only contain keys which exist in the Configstore DB  


### Using IAM Roles

You are able to specify an IAM Role when you set up a new Configstore:
```bash
configstore init --master-key "alias/my-key" --role "arn:aws:iam::123456789:role/someRole"
```
With an IAM Role specified, Configstore will assume that role whenever it needs to call the AWS API. There are however some
cases where you want to execute a Configstore command, but you do not want to assume the Role that is used when managing
the contents of the Configstore: one such example is wanting to make use of your Configstore on an EC2 Instance. Since EC2
Instances can't assume IAM Roles (they have Instance Roles instead), you'll need to bypass the extra "assume role" mechanism
in Configstore, and rely simply on the credentials available on the instance. You can do this via the `--ignore-role` flag:
```bash
configstore get --ignore-role mykey
```
> NOTE: Make sure that the Instance Role used by the EC2 Instance has permissions to use the KMS Key set for the Configstore
to encrypt/decrypt data.


### Insecure Mode

In some cases (for example local Dev environments), it's a pain to have to have AWS access whenever you want to read or write
values in the Configstore DB. To get around this, you can set the `--insecure` flag when calling `configstore init`.
In this mode, the encryption key is generated locally (rather than via AWS KMS), and then stored in
**plain text form** in the `configstore.json` file. This allows us to keep the internals mostly the same, while sacrificing
security.

> WARNING: This mode is NOT suitable for production use, since anyone who has access to the `configstore.json` file
will be able to decrypt the secrets stored within it!


## Development

### Prerequisites

First, you'll need to install Go (`>= 1.12`) - on Mac OS you can do this via Homebrew:
```
brew install go --with-cc-common
```
While it is no longer a requirement with Go `1.8` and later, you'll need to also set the `GOPATH` environment variable for
the [Glide](https://glide.sh/) dependency manager to work (as of version `0.12.3` - should be fixed in a future release).
By default this is `~/go`.

Next, you'll need to check this Git repo out in your Go workspace, under `$GOPATH/src/github.com/CultBeauty/configstore`.

> NOTE: The `github.com/CultBeauty/configstore` bit is important, otherwise Go will not be able to resolve the internal dependency between packages of the application

Finally, you need to install [Glide](https://glide.sh/), a dependency manager for Go packages. On Mac OS, you can do this via Homebrew:
```
brew install glide
```

### Fetching dependencies

Dependencies are defined inside `glide.yaml`, with installed versions locked down in `glide.lock`.
Run `glide install` inside the source root to fetch these dependencies.

### App Layout

The Configstore app is split into two packages:

1. The client library under `client/`, which contains all the logic for managing a Configstore database, encrypt/decrypt secrets and so on
2. The `configstore` CLI application under `cmd/configstore`, which is basically just a wrapper around the client library

### Building

Once all the above is done, you can use `./build.sh` to build the Mac OS, Linux and Windows versions under `bin/darwin/configstore`,
`bin/linux/configstore` and `bin/windows/configstore` respectively.

### Testing

The tests use a combination of the Go testing framework for exercising the internal client library, and the [BATS](https://github.com/sstephenson/bats)
application for black-box testing the commands in the built version of the app.

To run the tests you'll need to install BATS first:
```bash
brew install bats
```

Once done, just execute `./test.sh`, which will first run the Go tests, and then the BATS tests (defined in `configstore.bats`).


### Releasing new version

Once you built all the packages, you can use `./release.sh $version`, passing in the `$version` number you want to release.
If the version you're releasing doesn't match the version defined in `main.go`, the script will raise an error.
When run, the script does the following:

 1. Creates `tar.gz` archives with the version number included in the file name, for each platform separately
 2. Creates a Git Tag with the specified version number and `v` as a prefix, for example `v0.0.4`
 3. Pushes the Git Tag to GitHub
 
Once the release script has run, you'll want to create a "proper" release in GitHub, which includes the pre-built
binaries as well.
By default GitHub insist on listing every Tag as a release, but don't let this "feature" throw you off. Just do the following:

 1. Go to the repo on GitHub
 2. Click the "Releases" tab
 3. Click the "Draft a new release" button
 4. Type the tag you just created into "Tag version"
 5. Type the version number prefixed by "Release " into "Release Title" (for example "Release v0.0.4")
 6. You may want to type a changelog into "Description" for good measure
 7. Upload the archive files created by the release script above
 8. Click "Publish release"
