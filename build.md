#!/usr/bin/env rundown --default build

# Build simple <r label=build/>

Just a simple build of `rundown`:

``` bash reveal setup
go build -o rundown
```

# Build release <r label=release/>

This script builds the rundown release into `dist/`

The `.current-version` file is used as the basis of the generated version number, the build script increments the build number segment of it.

``` bash named stdout env
# Check for current version
export VERSION=`cat .current-version`
echo "This build version will be $VERSION"
```

``` bash named_all
# Building

if [ -d dist ]; then
  rm -rf dist
fi

mkdir -p dist/darwin-amd64/
mkdir -p dist/linux-amd64/

# Building MacOS
GOOS=darwin go build -o dist/darwin-amd64/rundown
# Building Linux
GOOS=linux go build -o dist/linux-amd64/rundown

# Preparing release
cp LICENSE bash_autocomplete README.md dist/darwin-amd64/
cp LICENSE bash_autocomplete README.md dist/linux-amd64/

tar -zcf dist/rundown-$VERSION-darwin-amd64.tgz dist/darwin-amd64
tar -zcf dist/rundown-$VERSION-linux-amd64.tgz dist/linux-amd64
```

Rundown built and available at `dist/rundown`.

## Setup local autocomplete <r label=autocomplete/>

Bash autocomplete can be added via:

``` bash reveal named
# Setup autocomplete
PROG=rundown source dist/darwin-amd64/bash_autocomplete
```

If installing via Homebrew or a package manager, this should be done for you.

## Increment Version <r label=incr/>

The `.current-version` file is used as the basis of the generated version number, the build script increments the build number segment of it.

``` ruby named stdout
# Increment version

vers = IO.read(".current-version").split(".")

vers[-1] = vers[-1].to_i + 1
vers = vers.join(".")

IO.write(".current-version", vers)

puts "Next build version will be #{vers}"
```

# Rundown Debugging Build <r label=debug/>

To debug with Delve, build Rundown with optimisations disabled:

``` bash reveal setup
go build -gcflags="all=-N -l" -o rundown
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

# Building a video

<r desc>Builds an animated gif video from a quicktime file.</r>

<!--~
``` bash stdout nospin
if [ -z "{$FILE:-}" ]; then
  echo "Specify a FILE env to run this"
  exit 1
fi
```
-->

``` bash named
# Generating video file
ffmpeg -ss 00:00:12.000 -i /Users/elseano/Desktop/Screen\ Recording\ 2020-09-27\ at\ 10.50.17\ am.mov  -pix_fmt rgb8 -r 10 screen.gif
convert -layers Optimize $FILENAME opt_$FILENAME.gif
```