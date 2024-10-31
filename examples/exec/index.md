# Running Commands in Rundown

The whole point of Rundown is to run code blocks within markdown files.

## Basic Usage

A `<r>` element immediately preceding a tagged code block will cause rundown to execute that block using the tagged command.

Example:

~~~ md
<r />

``` bash
ls -la
```
~~~

Internally, Rundown will search the `$PATH` for an executable named `bash`. It then generates a temporary script file which looks like this:

~~~ bash
#!/bin/bash

ls -la
~~~

So any command which can act as an interpreter can be run.

After execution, the temporary files are deleted.

## Customising the interpreter

In situations where the command can't act as an interpreter, or you need to customise how something is invoked, you can use the `with` tag on the rundown block.

Example:

~~~ md
<r with="mysql -h db-host < $SCRIPT_FILE | column -t" />

``` sql
SELECT * FROM Users;
```
~~~

In this case, Rundown creates two scripts. Script one (which is the script which rundown executes):

``` bash
mysql -h db-host < /path/of/script2 | column -t
```

And script 2, which is invoked from script 1:

``` sql
SELECT * FROM Users;
```

