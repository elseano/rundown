<r opt="docopt" type="string" desc="An option for the document"/>

# Rundown test file

This file tests rundown.


## Some heading <r label="greets"/>

<r desc="Greets you by your name"/>

<r opt="name" type="string" desc="The name to greet" required/>

<r stdout/>

``` bash
echo "Hello $OPT_NAME"
```


## Say Goodbye <r label="byee"/>

<r desc="Asks for your name, and then says goodbye"/>

<r opt="0:name" type="string" desc="The name to greet"/>
<r opt="*:misc_stuff" type="string" desc="Other names"/>

<r stdout/>

``` bash
echo "Bye $OPT_NAME"
echo "Cya $OPT_MISC"
```


## Release <r label="release"/>

<r spinner="Listing previous versions..." stdout/>

``` bash
git tag -l | tail -n 3
```

<r opt="version" as="VERSION" type="string" required desc="The version to release" prompt="Version to release" />

<r stdout/>

``` bash
git tag -a $VERSION -m "$VERSION"
goreleaser --rm-dist --skip-validate
```