# Sections

Sections in Rundown are denoted via `<r section="something"/>` placed inside headings. These are the core of how you build out a suite of commands for the CLI.

For example:

``` markdown
# Do a thing <r section="do-a-thing"/>

This is a thing you're doing.

# Do a different thing <r section="another-thing"/>

This is another thing.
```

These become subcommands when calling rundown, and are detailed within the `--help` flag:

```
$ rundown --help
Rundown turns Markdown files into console scripts.

Usage:
  rundown [command] [flags]...
  rundown [command]

Available Commands:
  help             Help about any command
  do-a-thing       Do a thing
  another-thing    Do a different thing
```

They can be invoked:

```
$ rundown do-a-thing
## Do a thing

This is a thing you're doing.
```

## Section Options/Flags

Sections can have flags which allows you to build out more advanced scripts:

~~~ markdown
# Some command <r section="something"/>

<r opt="edition" as="MY_EDITION" type="string" required desc="The edition to use" default="blank" />

<r spinner="Loading edition $MY_EDITION..." />

``` bash
curl https://someurl/download/$MY_EDITION
```
~~~

When invoking rundown's help, the option will be documented:

```
$ rundown something --help
Some command

Usage:
  rundown something [flags]

Available Commands:
  help             Help about any command

Flags:
  --edition string   The edition to use (default "blank")
```

Flags are also provided by rundown's shell autocompletion.

### Option types

Rundown supports the following `type` values for the `<r opt>` tag:

* `string` - Any string value
* `bool` - True/false. Presence of the flag indicates true.
* `enum` - Any of the provided values in the format of `type="enum:one|two|three"`
* `file` - A filename is expected.
  * `file:exist` - The filename must exist.
  * `file:not-exist` - The filename must not exist.

A note on the `file` type:

Because rundown executes all scripts relative to the rundown file containing the script, providing a filename using a `string` input type will result in a file path you didn't likely expect. 

For example, given this file tree:

```
|- foo/
    |- RUNDOWN.md
|- examples/
```

Running `rundown` from the `examples` directory and providing a filename to a `string` option will result that filename being interpreted relative to the `foo` directory by any scripts in that file. By using a `file` option type, rundown will know to repath the given filename to the invocation location.

## Ending a Section

A section ends when:

* Rundown reaches a new heading at the same level as the section.
* End of the document is reached.

This means that sub-headings can be both their own sections, and part of their parent section. When Rundown encounters this, it follows the rules of "Section Fall-through" and "Conditionals" as described below.

## Section Fall-through

Higher level headings are included as part of a section, and can also be considered their own section. For example, here running `do-a-thing` will also run `another-thing`. However you can also run `another-thing` by itself.

``` markdown
# Do a thing <r section="do-a-thing"/>

This is a thing you're doing.

## Do a different thing <r section="another-thing"/>

This is another thing.
```

# Branching & Conditionals <r section="branching"/>

Sometimes you need to branch the code rundown needs to run. While you can easily use `bash` scripting for this, sometimes you want a little more.

Sections support the `if` attribute, which allows everything under that heading to be skipped if the script evaluates to a non-zero result. For example:

``` markdown
# Do a thing

This is a thing

## Do a sub-thing <r if="false"/>

This won't be run.

## Do another sub-thing <r if="true"/>

This will be run.
```

Results in:

``` expected
# Do a thing

This is a thing

## Do another sub-thing

This will be run.
```

The `if` is only evaluated when falling through. It doesn't prevent the heading from being called directly as a command. For example, here `sub-thing` won't run as a fall through, but can be run directly.

``` markdown
## Do a sub-thing <r section="thing"/>

Running from the `thing` command.

### Do another sub-thing <r section="sub-thing" if="false"/>

This will only appear when running `sub-thing` directly.
```

## Ending the branch <r section="branching:ending"/>

In situations where you need to do either (A) or (B), and then always do (C), you can end a branch with an invisible heading. These also work embedded inside invisible blocks.

``` markdown
## Do a thing <r section="thing"/>

I'll be rendered.

### Don't do a thing <r if="false"/>

I won't be rendered.

### Do another thing <r if="true"/>

I will be rendered.

<!--~

###

I will be rendered whatever happens.

-->
```

Will result in:

``` expected
## Do a thing

I'll be rendered.

### Do another thing

I will be rendered.

I will be rendered whatever happens.
```

## Dependencies <r section="deps" />

Dependencies can be specified using the `dep` attribute. The same dependency encountered multiple times will only run once. For example:

~~~ markdown

# Run Me <r section="run" />

<r dep="dep1">I depend on 1.</r>

<r dep="dep2">I depend on 2.</r>

Now I'm doing my thing.

# Dependency 1 <r section="dep1" />

<r dep="dep3">I depend on 3.</r>

Dependency 1.

# Dependency 2 <r section="dep2" />

<r dep="dep3">I depend on 3.</r>

Dependency 2.

# Dependency 3 <r section="dep3" />

I'm dependency 3.

~~~

Running with `rundown run` will have the following flow:

1. First `run` is started, and the heading is written.
2. `dep1` is encountered, and begins running.
3. `dep3` is encountered inside `dep1`, and begins running.
4. `dep3` finishes, and `dep1` resumes.
5. `dep1` finishes, and `run` resumes.
6. `dep2` is encountered, and begins running.
7. `dep3` is encountered, but has been run already, so is skipped.
8. `dep2` finishes, and `run` resumes.

The output will appear as follows:

~~~ expected
# Run Me

I'm dependency 3.

Dependency 1.

Dependency 2.

Now I'm doing my thing.
~~~

Note that other sections won't have their headings displayed. This is because it can cause confusing output, where the contents of `run` would appear under the `dep2` heading.

## Invocations <r section="invokes"/>

Invocations operate the same as dependencies, however will always be run no matter how many times it's encountered. Because of this, invocations support options. For example:

~~~ markdown

# Run me <r section="run"/>

<r spinner="Calculating things..." capture-env="ARG" />

``` bash
ARG="Some arg"
```

The arg is: <r sub-env>$ARG</r>.

<r invoke="print-thing" text="$ARG">See [Printing](#printing) for an example on how to print</r>

# Printing <r section="print-thing" silent />

<r opt="text" as="TEXT" type="string" required />

The result is: <r sub-env>$TEXT</r>.

~~~

When run with `rundown run` results in:

~~~ expected
# Run me

✔ Calculating things...

The arg is: Some arg.

The result is: Some arg.
~~~