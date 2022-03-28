<r opt="docopt" type="string" desc="An option for the document"/>

# Rundown's Rundown File


## Release <r section="release"/>

<r help>

Creates a git tag for the provided version, and runs `go-releaser`.

</r>

<r opt="version" as="VERSION" required type="string" desc="The release version (i.e. v0.4.0-beta.6)"/>

<r named-all stdout/>

``` bash
# Tagging the release...
git tag -a $VERSION -m "First release"
git push origin $VERSION
source .env && goreleaser release --skip-validate --rm-dist
```


## Test no spin <r section="test:nospin"/>

<r nospin stdout/>

``` bash
echo "Hi"
```

## Test spinner change <r section="test:spin-change"/>

<r named-all stdout/>

``` bash
# Doing the hi
sleep 1
echo "Hi"

# Doing something else
sleep 1
echo "Bye"
```

## Test curl response <r section="test:curl"/>

<r stdout spinner="Requesting..."/>

``` bash
curl http://example.org
```

## Test curl response <r section="test:ls"/>

<r stdout spinner="Executing..." nospin/>

``` bash
ls -la --color=always
```


## Some heading <r label="test:greets"/>

<r desc="Greets you by your name"/>

<r opt="name" type="string" desc="The name to greet" required/>

<r opt="greet" type="enum|hi|formal" desc="The greeting style" required/>

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


## Environment <r section="test:env" />

<r capture-env spinner="Setting env..."/>

``` bash
export RESULT=one
```

The result is: <r sub-env>$RESULT</r>.