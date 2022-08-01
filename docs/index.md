# Rundown Documentation

By default, Rundown will execute a `RUNDOWN.md` file, looking in the current directory upwards until it's found.

All commands in a file are executed with the working directory being the same as that file. This allows you to run rundown commands from anywhere in a directory structure.

## Rundown flavoured Markdown

Rundown flavoured Markdown is designed to be ignored by markdown renderers, so a Rundown file should appear as a normal Markdown file when viewing it in popular platforms, such as GitHub, GitLab, etc. 

This is achevied by a non-existant HTML element of the form:

``` html
<r attr attr key="value" .../>
```

By far this is the most common form you'll be using. However there are cases where you need to use the block version of the element.

This can be either:

``` html
<r attr key="value">(some content)</r>
```

Or for more complex copy (**note** that the newlines are required as per the Markdown spec):

``` html
<r attr key="value" ...>
(newline)
Some content
(newline)
</r>
```

Anything within the block element will be rendered by Markdown renderers, however depending on the tag may not be shown in Rundown.

Finally, there's the rundown magic comment `<!--~ ... -->` (note the tilde `~`), which will hide content from Markdown renderers, but still be included in Rundown.

The following sections are both documentation and part of Rundown's automated test suite:

* <r import="run">[Running Code](./code.md)</r>
* <r import="sections">[Sections, Commands and Branching](./sections.md)</r>
* <r import="templating">[Templating](./templating.md)</r>
* <r import="stop">[Stopping scripts early](./stop.md)</r>
* [Importing](./importing.md)