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


## Environment <r section="env" />

<r capture-env spinner="Setting env..."/>

``` bash
export RESULT=one
```

The result is: <r sub-env>$RESULT</r>.