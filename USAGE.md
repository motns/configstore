# Usage

### Initial Setup

In order to use the config store, it first needs to be initialised. However, before you can do that, you need to create
an AWS KMS Master Key - [see Documentation](http://docs.aws.amazon.com/kms/latest/developerguide/create-keys.html).

You'll also need to grab a copy of the `configstore` binary for your platform from [Releases](https://github.com/motns/configstore/releases),
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
Like `get`, this will decrypt secret values before printing them out. You can also filter the output to only show keys
which include a certain piece of text:
```bash
configstore ls username
```

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


### Executing Shell Commands

As an extension to the template processing capability, you can also call shell commands via Configstore, with template
variables filled in before execution:
```bash
configstore exec "my_command {{.foo}} {{.bar}}"
```


### Overrides

Overrides are helpful in cases where you have a single Configstore for an environment, but you need two or more versions
of it with only minor differences. Instead of having to duplicate the entire Configstore DB, you can override one or more
keys via an override file, which is basically just a JSON file with a single, flat object in it, where both keys and values
are strings.

Overrides are supported for any command which reads or outputs data (`get`, `ls`, `process_template`, etc.) via the
`--override` flag. You can provide multiple override files by passing multiple instances of `--override /path/to/file.json`. 

The rules for overrides are:
 * When given multiple override files, they are processed left to right, merging them together into a single override that's used internally.
   Basically, if you pass three override files which each contain the same key, the last one will "win".
 * Overrides cannot contain any encrypted values, and therefore can't be used to override secrets. If you have different secret values,
   you should have different Configstore DBs instead.
 * Your override can only contain keys which exist in the Configstore DB; you can't use the override to append new keys.


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


### Autocomplete

There's built-in support for autocomplete via BASH and Zsh. You can enable this by copying the respective autocomplete
file (`bash_autocomplete` or `zsh_autocomplete`) bundled in the Zip file you downloaded, into your home folder,
and sourcing it in your profile config.