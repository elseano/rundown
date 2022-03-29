# Executing Code in Rundown

Markdown has great support for code blocks using the fenced code block notation. Rundown extends upon this, by actually running these code blocks.

~~~ markdown
Now, build your docker container:

``` bash
docker build . -t some_tag
```
~~~

By default, rundown won't run fenced code blocks, because they may be unsafe to run by default.

To enable running of these code blocks, you can add the `<r/>` attribute before the code block to mark it as runnable. The attribute must contain at least one of the execution attributes outlined further below.

## Execution Attributes

How Rundown runs your code can be modified using these attributes:

* `spinner` - Displays a spinner while the code is being executed. By default, the spinner will just display "Running..."
* `stdout` - Renders stdout and stderr as the command runs.
* `with` - Change the interpreter used by rundown to run the code block.
* `stdout-into` - Copy STDOUT into an environment variable for use later.
* `capture-env` - Capture the environment variable for use later.

### Example 1 - Spinner Customisation <r section="spinner" />

The spinner attribute allows you to customise the spinner text...

~~~ markdown
<r spinner="Tweaking widgets..." />

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

Rundown will attempt to use the syntax name as the executable. This works in many situations, for example with Ruby:

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

In situations where the syntax differs from the executable name, we can specify how to run the code block using the `with` attribute:

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

### Example 7 - Dynamic spinners <r-disabled section="dynamic-spinner"/>

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