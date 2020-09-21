# Rundown

Rundown is a terminal application which turns Markdown into executable code, rendering the contents into the console as it progresses.

As Rundown emphasises keeping the markdown still readable as a document, it's a great way to produce executable documentation. 
Some of the usecases rundown suits are:

* Automated setup guides
* Tutorials
* Code building scripts
* Simple programs

As we use Rundown, we're finding plenty of other ways you can use it!

## Feature Highlights & Examples

Rundown will run markdown perfectly fine. As your rundown files get more complex, you'll want to start using the rundown extensions. Rundown's extensions are transparent additions to the markdown format which aren't rendered by standard markdown renderers (i.e. Github, etc), meaning a reader won't even notice the additions. 

This file is a rundown file for example!

Some of the additions Rundown brings are:

* Shortcodes, which allow you to only run portions of a markdown file.
* Fenced code block execution and progress indicator manipulation.
* Optional failure handling, script skipping, and STDOUT presentation.
* First-class emoji support either via UTF characters, or :rocket: (`:rocket:`) syntax.
* Invisible blocks, which are only rendered inside markdown and ignored by web based renderers.
* Visually appealing console markdown rendering
* Shebang support, allowing you to make your markdown files executables (POSIX)
* Rundown files can be designed to execute top to bottom, or present a menu to execute only a single part.
* Fast - rundown is written in Go, and works on Linux, Windows and Mac.

[](label:show-code-blocks)
## Fenced Code Block Examples

By default, a fenced code block which doesn't specify a language will be rendered out.

    ```
    This is a simple fenced code block, it won't be executed.
    ```

```
This is a simple fenced code block, it won't be executed.
```

However, if you specify the syntax, then rundown will execute that file, and show a spinner as the execution progresses.

    ``` bash
    sleep 1
    ```

``` bash
sleep 1
```

In Rundown, you can change how the code block executes by adding additional flag and parameters to the syntax line.

    ``` bash stdout
    echo "Output Line 1"
    sleep 1
    echo "Output line 2"
    ```

``` bash stdout
echo "Output Line 1"
sleep 1
echo "Output line 2"
```

There are cases were the syntax doesn't match the executable, or you need to add flags to the executable. You can use the `with` flag on the code block. Here, we're also using the `named` flag, which assumes the first line is a comment which should be the title of the spinner.

    ``` js with:"node --jitless" stdout named
    // Hi from NodeJS
    console.log("Hi from NodeJS")
    ```

``` js with:"node --jitless" stdout named
// Hi from NodeJS
console.log("Hi from NodeJS")
```

A full list of the modifiers and examples can be found in the [Modifiers Example](./examples/mods.md) markdown file.

[](label:shortcodes)
## Shortcodes

Headings can have "shortcodes" attached to them, which allows that heading (and all child headings) to be executed specifically. Specifying a shortcode can be done either before the heading, or inside the heading itself.

    ## Shortcodes [](label:shortcodes)

    or

    [](label:shortcodes)
    ## Shortcodes

That heading can then be run via `rundown README.md shortcodes`. Bash/Fish/ZSH completition is available for shortcodes, as well as a shortcode subcommand which lists available shortcodes within a document.

Related to shortcodes is the `setup` flag. It's common to write instructions where every level 2 heading runs under the assumption that something from the parent heading has been done. The setup flag on a code block means that any shortcodes on child headings should also run the parent code blocks with the `setup` flag present.

For example:

    # Build Project

    Make sure you've set your architecture correctly.

    ``` bash env setup
    export GO_ARCH=linux
    ```

    [](label:compile)
    ## Compile

    ``` bash
    go build -o rundown
    ```

When you execute `rundown README.md compile`, rundown will first execute the parent heading's ("Build Project") `setup` code blocks.

More examples can be found in the [Shortcodes Example](./examples/shortcodes.md) markdown file.