# Rundown Help Integration

Rundown allows your markdown files to augment the `--help` flag.

    rundown README.md --help

When help is requested, it presents a list of possible shortcodes and their descriptions and options. You can provide a shortcode called `rundown:help`, which if present will be rendered prior to the built-in functionality.

Try it with this file, and you'll see the below section rendered as help.

## How to use this file <r label="rundown:help"/>

This file actually doesn't do anything except render it's contents.