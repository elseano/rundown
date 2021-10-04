<r opt="docopt" type="string" desc="An option for the document"/>

# Rundown test file

This file tests rundown.


## Some heading <r label="greets"/>

<r desc="Greets you by your name"/>

<r opt="name" type="string" desc="The name to greet" required/>

<r opt="greet" type="enum|hi|formal" desc="The greeting style" required/>

<r stdout/>

``` bash
# sleep 2
echo "$OPT_GREET $OPT_NAME"
# sleep 2
```


## Say Goodbye <r section="byee"/>

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

<r section="vanilla" desc="Thing">

# Heading 1

Heading contents.

## Heading 2

Hi there.

</r>