# Configstore

A simple command line application written in Go, used to store a mixture of plain-text and encrypted values, in a plain-text
JSON file. This JSON file is safe to commit to version control.

The encryption uses [AWS KMS](https://aws.amazon.com/kms/) as a starting point: A generated **Data Key** is used to
encrypt the secrets themselves (AES 256), which is then itself encrypted using a KMS **Master Key** and stored in the
JSON file alongside the values.
Whenever the plain text value of a secret needs to be loaded, the **Data Key** is decrypted via AWS KMS, and that key
is then used to decrypt the secret value. The decrypted **Data Key** is then discarded; it is never stored in plain form.

1. [Usage](USAGE.md)
2. [Development](DEVELOPMENT.md)
3. [Configstore Package](PACKAGE.md)