# Sections

Sections in Rundown are denoted via `<r section="something"/>` placed inside headings.

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

## Ending a Section

A section ends when:

* Rundown reaches a new heading at the same level as the section.
* End of the document is reached.

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

Headings support the `if` attribute, which allows everything under that heading to be skipped. For example:

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