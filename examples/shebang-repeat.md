#!/usr/bin/env rundown --ask-repeat

# Shebang Support

Markdown files with a rundown shebang can be executed directly.

``` bash reveal norun
#!/usr/bin/env rundown
```

For example, this file is a shebang script, so you can run it directly. It's using the `--ask-repeat` option, which will run inside an REPL-like environment, continally asking which shortcode to run and running it.

Headings without shortcodes aren't included in the menu unless they're parents of menus which do have labels.

## Shebang options

* `--default` - Sets the default shortcode to run if none is specified.
* `--ask` - Present the menu once, if no shortcode is provided.
* `--ask-repeat` - Like ask, but returns back to the menu after completion. Ctrl-C cancels.

# Investigate Servers <r label=i>

<r desc>This will investigate servers.</r>

``` bash stdout
echo "Server results"
```

# Login server <r label=l>

<r desc>This is the description of this action</r>

Logging you into the server. Type `exit` to return.

``` bash borg
sh
```

You won't see this.

