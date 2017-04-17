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
NOTE: Instead of taking the value as an argument, the app will prompt you to type
it in with a silent input for security.

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


### Using template files

The app also has the ability to take a template file, and fill in values from the Configstore DB. The app supports Go's
plain-text template format - you can read more about it [here](https://golang.org/pkg/text/template/).

A simple example would be to create a template file `application.conf` with the contents:
```
settingA = {{.foo}}
settingB = {{.bar}}
```
and the call the Configstore app with:
```
configstore process_template application.conf
```
The app then loads all values from the Configstore DB (decrypting secret values in the process), and then attempts
to process the template file, replacing `{{.foo}}` and `{{.bar}}` in our example with the values stored under those keys (`foo`, `bar`).
If you reference a key in your template file that doesn't exist in the Configstore, the app will raise an error.
The processed output is then sent to `stdout`, so you can just pipe it into another file for storage. 

### Insecure Mode

In some cases (for example local Dev environments), it's a pain to have to have AWS access whenever you want to read or write
values in the Configstore DB. To get around this, you can set the `--insecure` flag when calling `configstore init`.
In this mode, the encryption key is generated locally (rather than via AWS KMS), and then stored in
**plain text form** in the `configstore.json` file. This allows us to keep the internals mostly the same, while sacrificing
security.

> PLEASE NOTE that this mode is NOT suitable for production use, since anyone who has access to the `configstore.json` file
will be able to decrypt the secrets stored within it.


## Development

### Prerequisites

First, you'll need to install Go (duh!) - on Mac OS you can do this via Homebrew:
```
brew install go --with-cc-common
```
While it is no longer a requirement with Go `1.8` and later, you'll need to also set the `GOPATH` environment variable for
the [Glide](https://glide.sh/) dependency manager to work (as of version `0.12.3` - should be fixed in a future release).
By default this is `~/go`.

Next, you'll need to check this Git repo out in your Go workspace, under `$GOPATH/src/`.

> NOTE: Because of the way Go handles dependencies under `/vendor`, this project will only work inside `$GOPATH/src/`. Try to build it outside of that path, and you'll see a number of errors relating to missing dependencies.

Finally, you need to install [Glide](https://glide.sh/), a dependency manager for Go packages. On Mac OS, you can do this via Homebrew:
```
brew install glide
```

### Fetching dependencies

Dependencies are defined inside `glide.yaml`, with installed versions locked down in `glide.lock`.
Run `glide install` inside the source root to fetch these dependencies.


### Building

Once all the above is done, you can use `./build.sh` to build the Mac OS, Linux and Windows versions under `bin/darwin/configstore`,
`bin/linux/configstore` and `bin/windows/configstore` respectively.

### Releasing new version

Once you built all the packages, you can use `./release.sh $version`, passing in the `$version` number you want to release.
If the version you're releasing doesn't match the version defined in `main.go`, the script will raise an error.
When run, the script does the following:

 1. Creates `tar.gz` archives with the version number included in the file name, for each platform separately
 2. Creates a Git Tag with the specified version number
 3. Pushes the Git Tag to GitHub
 
Once the release script has run, you'll want to create a "proper" release in GitHub, which includes the pre-built
binaries as well.
By default GitHub insist on listing every Tag as a release, but don't let this "feature" throw you off. Just do the following:

 1. Go to the repo on GitHub
 2. Click the "Releases" tab
 3. Click the "Draft a new release" button
 4. Type the tag you just created into "Tag version"
 5. Type the version number with a `v` prefix into "Release Title" (for example "v0.0.4")
 6. You may want to type a changelog into "Description" for good measure
 7. Upload the archive files created by the release script above
 8. Click "Publish release"
