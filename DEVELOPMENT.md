# Development

### Prerequisites

First, you'll need to install Go (`>= 1.13`) - on Mac OS you can do this via Homebrew:
```
brew install go --with-cc-common
```
While it is no longer a requirement with Go `1.8` and later, you'll need to also set the `GOPATH` environment variable for
the [Glide](https://glide.sh/) dependency manager to work (as of version `0.12.3` - should be fixed in a future release).
By default this is `~/go`.

Next, you'll need to check this Git repo out in your Go workspace, under `$GOPATH/src/github.com/motns/configstore`.

> NOTE: The `github.com/motns/configstore` bit is important, otherwise Go will not be able to resolve the internal dependency between packages of the application

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
 3. Select "Tags"
 4. In the list of tags, you should see the one you just created. Click the menu (...) on the right hand-side, and click "Create release"
 5. Type the version number prefixed by "Release " into "Release Title" (for example "Release v2.0.0")
 6. You may want to type a changelog into "Description" for good measure
 7. Upload the archive files created by the release script above
 8. Click "Publish release"
