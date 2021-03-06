#!/usr/bin/env rundown --default build

# The Rundown Build File <r label="rundown:help"/>

![**Rundown**](logo.png)

This file contains all the essentials on building, testing and releasing Rundown.

By default, this script will just build a local copy of `rundown`. If you want to run the debugger, you'll want to run the `build:debug` or `delve` shortcodes.

# Build simple <r label=build/>

Just a simple build of `rundown`:

``` bash reveal setup env
go get -u github.com/mjibson/esc

export VERSION=`cat .current-version`
export GIT_COMMIT=$(git rev-list -1 HEAD)

FLAGS="-X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION"

go build -ldflags="$FLAGS" -o rundown cmd/rundown/main.go
```

Build version <r sub-env>**$VERSION**</r>.

# Build the Release <r label="release"/>

<r desc>Builds the release binaries and prepares all related assets</r>

## Build example files <r label=release:docs/>

<r desc>Takes the `_testdata/` markdown test files and copies the markdown section into the `docs/` folder.</r> This keeps documentation in sync with the tests.

``` bash spinner:"Generating man pages"
ronn --roff doc/*.ronn
```

## Build release binaries <r label=release:build/>

<r desc>This script builds the rundown release into `dist/`</r>

The `.current-version` file is used as the basis of the generated version number, the build script increments the build number segment of it.

``` bash named stdout env
# Check for current version
export VERSION=`cat .current-version`
echo "This build version will be $VERSION"
```

``` bash named-all
# Preparing to build

if [ -d dist ]; then
  rm -rf dist
fi

mkdir -p dist/darwin-amd64/
mkdir -p dist/linux-amd64/

GIT_COMMIT=$(git rev-list -1 HEAD)

FLAGS="-X main.GitCommit=$GIT_COMMIT -X main.Version=$VERSION"

# Building MacOS
GOOS=darwin go build -ldflags="$FLAGS" -o dist/darwin-amd64/rundown cmd/rundown/main.go
# Building Linux
GOOS=linux go build -ldflags="$FLAGS" -o dist/linux-amd64/rundown cmd/rundown/main.go

# Preparing release
cp LICENSE README.md logo.png dist/darwin-amd64/
cp LICENSE README.md logo.png dist/linux-amd64/

cp doc/rundown.1 dist/darwin-amd64/
cp doc/rundown.1 dist/linux-amd64/

mkdir dist/darwin-amd64/examples
mkdir dist/linux-amd64/examples

cp examples/*.md dist/darwin-amd64/examples
cp examples/*.md dist/linux-amd64/examples

# Creating platform archives...
cd dist/darwin-amd64 && tar -zcf ../rundown-$VERSION-darwin-amd64.tgz . && cd ../..
cd dist/linux-amd64 && tar -zcf ../rundown-$VERSION-linux-amd64.tgz . && cd ../..
```

Rundown built and available at `dist/rundown`.

## Increment Version <r label=release:incr/>

The `.current-version` file is used as the basis of the generated version number, the build script increments the build number segment of it.

``` ruby named stdout
# Increment version

vers = IO.read(".current-version").split(".")

vers[-1] = vers[-1].to_i + 1
vers = vers.join(".")

IO.write(".current-version", vers)

puts "Next build version will be #{vers}"
```

## Building a video <r label=release:video/>

<r desc>Builds an animated gif video from a quicktime file.</r>
<r opt="file" type="string" desc="The input QuickTime file"/>

<!--~
``` bash skip-on-failure stdout
if [ -z "{$FILE:-}" ]; then
  echo "Specify +file option to run this"
  exit 1
fi
```
-->

``` bash named
# Generating video file
ffmpeg -ss 00:00:12.000 -i /Users/elseano/Desktop/Screen\ Recording\ 2020-09-27\ at\ 10.50.17\ am.mov  -pix_fmt rgb8 -r 10 screen.gif
convert -layers Optimize $OPT_FILE opt_$OPT_FILE.gif
```

# Install Rundown Locally <r label="install"/>

<r desc>Installs rundown locally into the provided prefix.</r>

## Install local autocomplete <r label=install:autocomplete/>

Bash autocomplete can be added via:

``` bash reveal named
# Setup autocomplete
PROG=rundown source dist/darwin-amd64/bash_autocomplete
```

If installing via Homebrew or a package manager, this should be done for you.


# Rundown Debugging Build <r label=build:debug/>

To debug with Delve, build Rundown with optimisations disabled:

``` bash reveal setup
go build -gcflags="all=-N -l" -o rundown cmd/rundown/main.go
```

<r stop-ok comment="Don't continue into the Delve process by default, as it's hard to exit." />

## Delve <r label=delve/>

<r desc="Starts a delve remote debugging process"/>

Then start a delve session in this console. Connect to the session on `localhost:49491`.

To abort this process, run `killall dlv` in another shell, or disconnect in Visual Studio (once connected).

Debugging `rundown`'s handling of the file `debug.md`:

``` bash borg reveal
~/go/bin/dlv exec --api-version 2 --headless --listen 127.0.0.1:49491 ./rundown -- debug.md
```

<r stop-ok />

## Delve Run Test <r label=delve:test/>

Same as **Delve** above, but runs the specified test.

``` bash borg reveal
~/go/bin/dlv test --api-version 2 --headless --listen 127.0.0.1:49491 github.com/elseano/rundown -- -run TestHidden
```

# Testing Rundown

## Testing locally <r label=test/>

Simply run:

``` bash
go test ./...
```

## Testing other platforms <r label=test:all/>

<r desc>Rundown was built on OSX, and testing on Linux is done through a docker container.</r>

``` bash named
# Building docker container
docker build -t rdlinux -f build/Dockerfile.ubuntu .
```

Now that the container is ready, run the tests.

``` bash named stdout
# Running test inside Docker
docker run rdlinux
```
