# Rundown Automation

Time to get fancy.

While Rundown will execute any Markdown file top-to-bottom, soon you'll want to tailor how executions are displayed, tweak how a code block is actually run, or jump around the file a little bit.

Some people have even replaced simple Makefiles with a Rundown file. I probably wouldn't recommend this for complex Makefiles because there's no support for dependency graphs in Rundown.

Any-hoo.

## The Rundown Automation Element

The rundown automation element is a HTML element of the form:

``` html
<r attr attr key="value" .../>
```

By far this is the most common form you'll be using. However there are cases where you need to use the block version of the element.

This can be either:

``` html
<r attr key="value">(some content)</r>
```

Or, the more complex

``` html
<r attr key="value" ...>
(newline)
Some content
(newline)
</r>
```

Note that the newlines are required as per the Markdown spec. Luckily, this form is a rare requirement.

Note also, anything within the block element will be rendered to readers as well.

## Controlling the Spinner

* `<r named/>`

Treats the first line in the following code block as the spinner title.

``` bash
# This will become the spinner title
make release
```

* `<r nospin/>`

Hide the spinner, and display nothing when the block is executing. Can be handy for quick pre-flight checks when combined with the [Rundown Only Block](./rundown_only_block.md).

* `<r spinner="This will be the spinner title"/>`

Sets the spinner title directly. Handy when the `named` version doesn't make a lot of sense for readers.

## Controlling the Output

By default, the output from a running script is hidden.

* `<r stdout/>`

Displays STDOUT when executing the script.

* `<r stderr/>`

Displays STDERR when executing the script.

* `<r reveal/>`

By default, Rundown won't show the code block being executed. This reveals the code block, complete with syntax highlighting.

## Controlling code block execution

Under normal circumstances, Rundown will take the block of code, convert it into a temporary file, and execute it using the syntax as the name of the program.

For example:

~~~ markdown
``` ruby
puts "Hi"
```
~~~

Rundown will convert this into a [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) and run the resulting file:

``` ruby
#!/bin/env ruby
puts "Hi"
```

### Skip Execution

Forms:

* `<r norun/>`

Don't run the code block.

### Custom Executable Invocation

Forms:

* `<r with="executable"/>`

This changes the executable used for the file, and is required when the syntax doesn't match the executable name, or if you want to add additional arguments to the executable. For example:

~~~ markdown
<r with="perl -f"/>

``` perl
some_perl_stuff()
```
~~~

This becomes:

``` perl
#!/bin/env perl -f

some_perl_stuff()
```

### Advanced Invocation

Forms:

* `<r with="executable < $FILE"/>` (i.e. Contents must be piped via STDIN)
* `<r with="executable $FILE --some-flag">` (i.e. Flags required after filename)
* Basically anything with `$FILE`.

This changes how the code block is fed into the executable. `$FILE` is a special variable only available within this context, and becomes the temporary file upon execution. For example:

~~~ markdown
<r with="mysql < $FILE"/>

``` sql
INSERT INTO things(name) VALUES 'Rundown';
```
~~~

This style is great when executables:

* Expect input to be piped in
* Require a different form of shebang (i.e. Go requires a leading `/` instead of `#`)

Another example:

~~~ markdown
<r with="go run $FILE"/>

``` go
package main

import "fmt"

func main() {
    fmt.Println("Hello World!")
}
```
~~~

### Save Content

Forms:

* `<r save="key"/>`

Instead of running the program, this saves the contents into a temporary file. The full path of that file is then available to subsequent running blocks via the `$KEY` environment variable.

### Save Content with Extension

Forms:

* `<r save="key.json"/>`

As above, but sets the file extension on the temporary file, which is sometimes required.

### Environment Substitution

Forms:

* `<r sub-env/>`

Performs preprocessing on the file before execution, replacing `$NAME` with the contents of `NAME` within the environment.

This also works as a content tag when rendering Markdown. For example:

``` markdown
Your files have been installed into <r sub-env>$PREFIX</r>.
```

### Capture Environment

Forms:

* `<r env/>`

Captures environment variables, making them available to subsequent code blocks. All environment variables are captured.

Note: This only works for `bash` scripts.

If you need to do this from within something other than a shell script (i.e. from a Ruby script), then you can access the Rundown RPC interface directly:

``` ruby
IO.write(ENV["RUNDOWN"], "env: SOMETHING=Value\n")
```

Better support for non-bash and non-shell is coming.

### Borg Mode

Forms:

* `<r borg/>`

Borg mode makes the script assimilate the Rundown process, ending the Rundown execution and replacing it with the exection of the given code block.

This is handy for things like logging into servers via SSH. For example:

~~~ markdown
Logging you into Production...

<r borg/> 

``` bash
ssh -I $HOME/.keys/production.pem ec2-user@www.google.com
```
~~~

## Controlling Rundown Execution

By default, Rundown will execute your Markdown file from top-to-bottom. However, sometimes you'd rather provide something more like a bunch of handy shortcuts rather than a sequential execution.

### Creating Sections

Forms:

* `# Heading <r label="name"/>` 

Defines a section within your Markdown file.

Sections allow you to create subcommands within your markdown file, by defining the section within a heading. 

For example:

``` markdown
## Configure System <r label="system:config"/>
```

This heading can be invoked directly via `rundown FILE.md system:config`.

Additionally, running `rundown FILE.md --help` will list all the sections defined.

#### Naming Practice

Generally, I've found it easier when section names follow a column separated format based on the heading level, however Rundown does not enforce any conventions on you.

``` markdown

# Heading <r label="something"/>

## More Stuff <r label="something:more"/>

### More Specific <r label="something:more:specific"/>

```


### Ignoring Content

Forms:

* `<r ignore>...</r>`

One of the rare cases you'll use a block tag, instructs Rundown to ignore (don't render, don't execute) the block's contents.

### Describing Sections

Forms:

* `<r desc>...</r>`
* `<r desc="..."/>`

Provides a description for whatever heading the content is under, or for the document if it's before the first heading. The description is presented alongside the section's label when using the `--help` flag.

Using the first form makes the description also available to readers (will be rendered by web browsers), while the second form only shows the description when using `--help`.


### Providing Section Options (Parameters)

Forms:

* `<r opt="name" type="string" [required] [default="default value"] desc="Desc"/>`

Defines a option for your section, allowing sections to take arguments when they're invoked. These options are displayed alongside the section when using `--help`.

The provided option's value (or the default value if none provided) is available throughout the Section in the environment variable `$OPT_<NAME>`.

#### Valid types

* `string` - A free-form string.
* `int` - A number.
* `enum|opt1|opt2|...` - Must be a string `opt1` or `opt2`.
* `password` - Masked password input.

#### Renaming the environment variable

The alternative form:

* `<r opt="name" as="VAR" .../>`

Will instruct the option to put it's value into the environment variable `$VAR`.

#### Prompting for input

The alternate form:

* `<r opt="name" prompt="Enter password"/>`

Will pause rundown execution at the point where the option is defined, and ask the user for input of the prompted argument, _if_ the environment variable (or the `as` variable) isn't set.

### Prerequisite Blocks

Forms:

* `<r setup/>`

Defines this code block as a requirement for any section deeper than the current heading level. When directly invoking a child section, this code block will also be run. When invoking multiple child sections, the block will be run only once, on the first section.

#### Recommended usage of setup blocks

Structure your setup blocks to ensure that subsequent scripts fail for the right reason. For example, you probably don't want to install anything in a setup script, but you might want to verify something is installed.


### Halting Execution

Forms:

* `<r stop-fail/>`
* `<r stop-fail="Reason"/>`

Causes the script to exit with a failure status. The section form provides the ability to provide a message to the user.

Forms:

* `<r stop-ok/>`
* `<r stop-ok="Reason"/>`

Same as `stop-fail` but we're exiting with a success status.

### Skipping Execution

Forms:

* `<r skip-on-failure/>`

If the subsequent code block exists with a non-zero exit code, then skip to the next heading (any level). Otherwise, continue sequentially.

Forms:

* `<r skip-on-success/>`

If the subsequent code block exists with a zero exit code, then skip to the next heading. Otherwise, continue sequentially.

### Importing contents of another document

Rundown will by default attempt to run `RUNDOWN.md` falling back to `README.md`, unless you specify a filename to run.

If your markdown files are getting very long, you can extract out the automation part into another file, or files, and import them into the primary file, keeping things clean.

Forms:

* `<r import="scripts/another_file.md"/>`

The import path is always relative to the executing rundown file. The contents of `scripts/another_file.md` will be inserted into the current file at the point of import.

### Tying it all together

Lets take a rather extreme example of taking a simple task and making it runnable via Rundown with some options.

~~~ markdown
# Reseed Database

To wipe and reseed your development database, run the following SQL:

``` sql
DROP DATABASE IF EXISTS dev_db;
CREATE DATABASE dev_db;
```

Then, run migrations:

``` bash
rake db:migrate db:seed
```

~~~

To turn this into a friendly section which can be invoked, we use the following automation tags:

* `label` - Defines the label of the section.
* `desc` - Provides a description of the section.
* `ignore` - Hides the content when running inside Rundown.

Additionally, we'll add some other tags mentioned earlier.

~~~ markdown
# Reseed Database <r label="db:reseed"/>

<r desc="Wipes and reseeds the database" ignore>To wipe and reseed your development database, run the following SQL:</r>
<r opt="name" type="string" default="dev_db" desc="The name of the Database"/>

## Recreate Database

<r ignore>If you want to start from scratch, first run this SQL:</r>

<r spinner="Recreating database..." with="sed 's/dev_db/$OPT_NAME/' $FILE | mysql'/>

``` sql 
DROP DATABASE IF EXISTS dev_db;
CREATE DATABASE dev_db;
```

## Migrate and Seed database

Then, run migrations:

<r spinner="Migrating and seeding DB..."/>

``` bash
rake db:migrate db:seed
```
~~~

We can then invoke this file as follows:

* `rundown TASKS.md db:reseed` Default execution, reseeds database `dev_db`.
* `rundown TASKS.md db:reseed +name=prod_db` Use the database named `prod_db` instead.

## Handling Failure

When Rundown encounters a script which errors (returns a non-zero exit code), it will abort running the entire script, display the script with the error line highlighted, and the error message.

```
  ## Dependencies                     

  ✖ Installing Homebrew (Failed)

Error - exit status 1 in:

  1: #!/usr/bin/env bash
  2: 
  3: set -Eeuo pipefail
  4: 
  5: # Installing Homebrew
  6: which brew || /bin/bsh -c "$(curl -fsSL https://.../install.sh)"

SCRIPT: line 6: /bin/bsh: No such file or directory
```

When you're using `skip-on-failure` and `skip-on-success`, errors cases instead change the control flow, so you won't see an error message.

You can also customise what rundown says to the user based on the contents of the error message, by using the `on-failure` handler.

Form:

* `<r on-failure="regular-expression">(content)</r>`

These handlers are inherited according to the heading levels. So document level handlers will be used throughout the document, but handlers defined on a 2nd-level heading will only apply to that heading and sub-headings.

When the error matches the regular expression, then the tag's contents are shown after the end of the above output. Any active handler which matches will be shown in the order they appear in the document.

## Adding Help to your File

Rundown provides integrated help for your Markdown file. Given the following file:

~~~ markdown
# Wait for Activity Complete <r label="wait-complete"/>

<r desc>Waits for something to be complete, while showing progress</r>
<r opt="name" type="string" required desc="The name of the object to wait for"/>
<r opt="status" type="string" default="complete" desc="The status to wait for"/>
<r opt="type" type="string" default="any" desc="The type of object"/>

* Name: <r sub-env>$OPT_NAME</r>
* Status: <r sub-env>$OPT_STATUS</r>
* Type: <r sub-env>$OPT_TYPE</r>

Done.
~~~

The shortcode options are documented:

```
$ rundown README.md --help

Supported options for README.md

  wait-complete         Waits for something to be complete, while showing progress                    
    +name=[string]      The name of the object to wait for (required)
    +status=[string]    The status to wait for (default: complete)
    +type=[string]      The type of object (default: any)
```

### Help Preface

In addition to documenting the Sections and Options of a file, you can also provide a help preface. This is done via the special Section called `rundown:help`:

~~~ markdown
## Using this File <r label="rundown:help"/>

This file handles all the tasks related to updating the widgets.

The most common task you're probably looking for is `rundown widgets.md update` which just updates everything.

~~~

When invoked via `rundown widgets.md --help`:

```
$ rundown widgets.md --help

   Rundown usage         
                        
  This file handles all the tasks related to updating the widgets.
                                       
  The most common task you're probably looking for is  rundown widgets.md update  which 
  just updates everything.


   Supported options for widgets.md


  update                Runs all updates
  update:one            Updates only one widget
```
