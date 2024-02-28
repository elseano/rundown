<r opt="docopt" type="string" desc="An option for the document"/>

# Rundown's Rundown File

## Check Current Ref <r section="check-ref" silent />

Checks the current ref. This should appear twice.

## Never run <r section="never-run" silent if="false" />

I should not appear.

## Invoke <r section="invoke"/>

Invokes another function.

<r invoke="never-run" />

<r invoke="check-ref" />

<r dep="check-ref">You won't see me in the output</r>

<r dep="check-ref" />

<r dep="check-ref" />

Done.

## Release <r section="release"/>

<r help>Creates a git tag for the provided version, and runs `go-releaser`.</r>

<r opt="version" as="VERSION" required type="string" desc="The release version (i.e. v0.4.0-beta.6)"/>

<r spinner="Tagging..." stdout>Go releaser works from git tags. So make sure there's a tag.</r>

``` bash
git tag -a $VERSION -m "First release"
git push origin $VERSION
```

<r spinner="Releasing $VERSION..." stdout>Then, run `goreleaser` to cross-compile and publish the release to GitHub</r>

``` bash
source .env && goreleaser release --skip-validate --rm-dist
```

## Test env spinner <r section="test:envspin"/>

<r spinner="Setting env" capture-env="NAME"/>

``` bash
NAME="Hi there"
```

<r spinner="Greeting $NAME" stdout/>

``` bash
echo "Hi"
```

## Test no spin <r section="test:nospin"/>

<r nospin stdout/>

``` bash
echo "Hi"
```

## Test spinner change <r section="test:spin-change"/>

<r spinner="Running..." named-all stdout />

``` bash
# Doing the hi
sleep 1
echo "Hi"


# Doing something else
sleep 1
echo "Bye"

# And again
sleep 1
echo "Cya!"
```

## Test curl response <r section="test:curl"/>

Makes a call to `http://example.org` via `curl` and renders the result as it is received.

<r stdout spinner="Requesting..."/>

``` bash
curl http://example.org
```

:rocket: Test completed successfully.

### Do seomthing else <r if="false" section="test:curl:done"/>

Blah

## Test curl response <r section="test:ls"/>

<r stdout spinner="Executing..." nospin/>

``` bash
ls -la --color=always
```

## Test Stop Ok <r section="test:stopok"/>

Renders.

<r stop-ok>Stopped.</r>

Doesn't render.

## Test Stop Ok If is True <r section="test:stopokift"/>

Renders.

<r stop-ok if="true">Stopped.</r>

Doesn't render.

## Test Stop Ok If is False <r section="test:stopokiff"/>

Renders.

<r stop-ok if="false">Doesn't Render.</r>

End naturally.

## Test Stop Fail <r section="test:stopfail"/>

Renders.

<r stop-fail>Stopped.</r>

Doesn't render.

## Test Stop Fail If is True <r section="test:stopfailift"/>

Renders.

<r stop-fail if="true">Stopped.</r>

Doesn't render.

## Test Stop Fail If is False <r section="test:stopfailiff"/>

Renders.

<r stop-fail if="false">Doesn't Render.</r>

End naturally.

## Some heading <r label="test:greets"/>

<r desc>Greets you by your name</r>

<r opt="name" type="string" desc="The name to greet" required/>

<r opt="greet" type="enum:hi|formal" desc="The greeting style" required/>

<r stdout/>

``` bash
# sleep 2
echo "$OPT_GREET $OPT_NAME"
# sleep 2
```


## Say Goodbye <r section="test:byee"/>

<r desc="Asks for your name, and then says goodbye, like a boss"/>

<r opt="0:name" type="string" desc="The name to greet"/>
<r opt="*:misc_stuff" type="string" desc="Other names"/>

<r reveal named-all/>

``` bash
# Bye
sleep 1
echo "Bye $OPT_NAME"
# Cya
sleep 1
echo "Cya $OPT_MISC"
# Done
sleep 1
```

## Say Goodbye <r section="test:subs"/>

Stuff.

<r spinner="Working..." sub-spinners/>

``` bash
#> Bye
sleep 1
echo "Bye"
#> Cya
echo "Cya"
#> Again
ls
#> Done
sleep 1
```


## Environment <r section="test:env" />

<r capture-env spinner="Setting env..."/>

``` bash
export RESULT=one
```

The result is: <r sub-env>$RESULT</r>.


## Borg <r section="test:borg" />

<r borg/>

``` bash
echo "Hello from borg process"
```

## Spinners <r section="test:spinners" />

<r capture-env="NAME,COMPLEX_NAME45" spinner="Setting env..." />

``` bash
export NAME=Thingo
export COMPLEX_NAME45=More Thingos
```

<r spinner="Running thing for $NAME..." stdout />

``` bash
echo "NAME is: $NAME"
```

<r spinner="Running thing for ${COMPLEX_NAME45}..." sub-spinners stdout />

``` bash
#> Simple test...
echo $COMPLEX_NAME45

#> Complex test $NAME...
echo $NAME
```

## Failure <r section="test:fail" />

<r spinner="SomeScript" />

``` bash
idontexit
```