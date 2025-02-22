# Executing Code in Rundown

The most common thing you'll be doing in Rundown is running code blocks. Markdown has great support for code blocks using the fenced code block notation. Rundown extends upon this, by actually running these code blocks. 

~~~ markdown
Now, build your docker container:

``` bash
docker build . -t some_tag
```
~~~

By default, rundown won't run fenced code blocks, because they may be unsafe to run by default.

To enable running of these code blocks, you can add the `<r/>` attribute before the code block to mark it as runnable. The attribute must contain at least one of the execution attributes outlined further below.

In the simplest form, a spinner name can be provided. In this example, rundown also won't render the contents of the `<r>` tag.

~~~ markdown
<r spinner="Building docker image...">Now, build your docker container:</r>

``` bash
docker build . -t some_tag
```
~~~


## Execution Attributes

How Rundown runs your code can be modified using these attributes:

* `spinner` - Displays a spinner while the code is being executed. By default, the spinner will just display "Running..."
* `stdout` - Renders stdout and stderr as the command runs.
* `with` - Change the interpreter used by rundown to run the code block.
* `stdout-into` - Copy STDOUT into an environment variable for use later.
* `capture-env` - Capture the specified environment variables for use later.
* `if` - Only run the code block if the if script has an exit code of zero.
* `replace` - Perform a simple find/replace in the script, providing rudimentary templating.

### Example 1 - Spinner Customisation <r section="spinner" />

The spinner attribute allows you to customise the spinner text...

~~~ markdown
<r spinner="Tweaking widgets...">To tweak your widgets, do this:</r>

``` bash
sleep 1
```
~~~

Rundown will render this:

~~~ expected
✔ Tweaking widgets...
~~~

### Example 2 - Showing stdout <r section="stdout"/>

Given this:

~~~ markdown
<r spinner="Launching rocket..." stdout />

``` bash
echo "T-1..."
sleep 1
echo "T-0..."
```
~~~

Rundown will render this:

~~~ expected
↓ Launching rocket...
    T-1...
    T-0...
✔ Launching rocket...
~~~

### Example 3 - Using stdout-into with sub-env <r section="into"/>

The `stdout-into` copies stdout into an environment variable. This can be referenced later, either in other scripts, or in a `sub-env` command:

~~~ markdown
<r spinner="Doing stuff..." stdout-into="RESULT" />

``` bash
echo "Hi there!"
```

<r stdout/>

``` bash
echo $RESULT
```

We just wrote <r sub-env>$RESULT</r>. Did you see it?
~~~


Should result in the following output:

``` expected
✔ Doing stuff...
↓ Running...
    Hi there!
✔ Running...

We just wrote Hi there!. Did you see it?
```

### Example 4 - More than just shell scripts <r section="int"/>

Rundown will attempt to use the provided syntax name as the executable. This works in many situations, for example with Ruby:

~~~ markdown
We're going to run Ruby:

<r stdout/>

``` ruby
puts "Hi from Ruby!"
```
~~~

As with bash, this will result in:

~~~ expected
We're going to run Ruby:

↓ Running...
    Hi from Ruby!
✔ Running...
~~~

### Example 5 - Custom interpreter <r section="int-custom"/>

In situations where the syntax differs from the executable name, we can specify how to run the code block using the `with` attribute. This attribute accepts anything you can do in bash, such as chaining commands via pipes.

When doing this, `$SCRIPT_FILE` is a special environment variable provided by rundown which contains the absolute path to a temporary file containing the script provided.

~~~ markdown
SQL can't be run directly. Lets just write it out to the console:

<r stdout with="cat $SCRIPT_FILE" spinner="Querying database..."/>

``` sql
SELECT * FROM Users
```
~~~

As with bash, this will result in:

~~~ expected
SQL can't be run directly. Lets just write it out to the console:

↓ Querying database...
    SELECT * FROM Users
✔ Querying database...
~~~

### Example 6 - A broken script <r-disabled section="broken"/>

There's going to be times when your script won't work. Rundown does it's best to highlight where your script broke:

~~~ markdown
<r spinner="Running..." />

``` bash
echo "Hi there"
this_command_doesnt_exist
echo "Never reached"
```
~~~

Should render:

``` expected-err
✖ Running...

Script Failed:
       1: echo "Hi there"
  *    2: this_command_doesnt_exist
       3: echo "Never reached"
  
  Line 2: this_command_doesnt_exist: command not found
```

## Spinners

By default, the spinner message is set as "Running...", however this can be changed a few ways.


### Renaming the Spinner <r section="spinner:rename"/>

The most common is to provide new spinner text using the `spinner` attribute.

~~~ markdown
<r spinner="Doing stuff..." />

``` bash
echo "Doing some stuff..."
```
~~~

Will render:

~~~ expected
✔ Doing stuff...
~~~

### Multi-step spinners <r section="spinner:multi"/>

When using `bash` or `sh` scripts, you can turn comments into spinners. Each comment creates a new sub-spinner.

Comments must start with `#> ` instead of `#` to differentiate between regular comments and multi-step commands.

For example:

~~~ markdown
<r spinner="Building..." sub-spinners />

``` bash
#> Beating eggs
echo "Beat the eggs"

#> Cooking eggs
echo "Cook the eggs"

#> Eating omlette
echo "Eat delicious omlette"
~~~

Creates nested spinners:

``` expected
- Building...
  ✔ Beating eggs
  ✔ Cooking eggs
  ✔ Eating omlette
✔ Building...
```

When using something other than bash, you can still support nested spinners. Rundown looks for an [ANSI OSC](https://en.wikipedia.org/wiki/ANSI_escape_code#OSC_(Operating_System_Command)_sequences) command in `stdout` of the format `ESC ] R;SETSPINNER (Base64 Encoded Value) BEL`.

### Dynamic spinners <r-disabled section="spinner:dynamic"/>

Sometimes you'd like to have your spinners be a bit more descriptive. Rundown expands environment variables in spinner names:

~~~ markdown
<r capture-env="TITLE"/>

``` bash
TITLE="Some title"
```

<r spinner="$TITLE..." stdout/>

``` bash
echo "Hi from a dynamic spinner"
```
~~~

Will render:

~~~ expected
✔ Running...
↓ Some title...
    Hi from a dynamic spinner
✔ Some title...
~~~

## Hidden code <r section="hidden"/>

This should be used rarely.

~~~ markdown
I will be read.

<!--~
I will only appear in Rundown.

<r spinner="Running"/>

``` bash
echo "Hi"
```
-->
~~~

Results in:

``` expected
I will be read.

I will only appear in Rundown.

✔ Running
```

## Propogating Environment Variables

Environment variables can be captured from scripts, and made available to both Rundown and subsequent scripts by using the `capture-env` attribute.

### Using variable inside rundown <r section="env:markdown"/>

Once a script has been run, any captured variable can be rendered in the markdown content by surrounding it with a `<r sub-env>` element.

~~~ markdown
<r capture-env="GREETING" />

``` bash
GREETING="Hi there"
```

The greeting is: <r sub-env>$GREETING</r>.
~~~

Will result in:

~~~ expected
✔ Running...

The greeting is: Hi there.
~~~

You can also set environment variables programmatically using an ANSI OSC command `ESC ] R;SETENV KEY=VALUE BEL` written to `stdout`.

### Saving the current working directory <r section="env:pwd" />

All captured environment variables are provided to subsequent execution blocks. 

A special case is the `PWD` variable, which can be used to set the working directory for subsequent scripts. By default, the initial `PWD` is always the directory containing the currently executing Rundown file. 

For example:

~~~ markdown

First, change to the desired directory:

<r capture-env="PWD"/>

``` bash
cd ..
```

Then see whats in it:

<r stdout/>

``` bash
ls -x1 docs | grep index
```

~~~

Should result in:

~~~ expected
First, change to the desired directory:

✔ Running...

Then see whats in it:

↓ Running...
    index.md
✔ Running...

~~~
