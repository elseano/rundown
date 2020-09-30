# Modifier formats

Modifiers come in two flavours: Flags, and Values. Flag modifiers are single `words`, which simply imply they're activated. Values are a `key:value` format, and the value can optionally be quoted with either single or double quotes if it contains spaces.

Modifiers can be specified using one of three possible formats. Markdown renderers tend to vary, and in order to keep a document readable, you should choose the right format for your rendering situation.

## Tag Format

The most compatible way of specifying modifiers is using the `rundown` tag. It comes in two flavours:

* Short, which is just `<r flag key=value/>`
* Long, which is `<rundown flag key=value/>`

    <rundown nospin blah=something />

Browsers will ignore this tag, but Rundown will process it. If you want to do some funky stuff with the tag in the browser, you can use the Custom Elements API.

The rundown tag can appear before block level elements (paragraphs, headings, code blocks) and will alter them.

    <r nospin/>
    ``` bash
    # No spinner will be shown
    ```

You can also use the rundown tag to alter how text renders.

    Your user ID is <r sub-env>`$USER`</r>.

## Fenced code block shortcut

This is a alternative to the Tag Format specifically for fenced code blocks, and has limited support in markdown renderers.

* **GitHub** :check_mark_button:
* **GitLab** :cross_mark: Breaks syntax highlighting.
* **VS Code** :check_mark_button:

    ``` bash nospin
    sleep 1
    ```

# Code Modifiers

Modifiers allow you to change how the code executes.

## Blank <r label=blank/>

Default execution

    ``` bash
    sleep 1
    ```

## No Spinner <r label=nospin />

The `nospin` flag prevents the spinner from showing at all. Generally you'd combine this with the `stdout` flag below. Reach for this if scripts ask for input on a _blank_ line, which can cause the spinner to over-write any user input.

    ``` bash nospin
    sleep 1
    ```

## Named - CHanging the spinner title <r label=named/>

The `named` flag treats the first line of the script as a comment, and that comment body is the title of the spinner.

    ``` bash named
    # Waiting...

    sleep 1
    ```

Any number of applications can be running your script, so the comment line ignores anything that isn't a letter or number at the start of the line.

Take this Go script for example:

    ``` go named norun reveal
    // This is the spinner title
    printf("Oh hai")
    ```

## Stdout <r label=stdout/>

This shows STDOUT in the output.

    ``` bash
    echo "You won't see this"
    ```

    ``` bash stdout
    echo "You should see this"
    ```

## No Run <r label=norun/>

This ignores the code block, don't run it, and don't display it.

    ``` bash norun
    echo "This will be displayed and not run"
    ```

## Reveal <r label=reveal/>

This will display the code block as part of Rundown's output, **AND** run it immediately afterwards.

    ``` bash reveal
    sleep 1
    ```

A common approach to rendering syntax highlighted code in the console is to use `norun` with `reveal`:

    ``` ruby norun reveal
    def method()
      "Hi, I'm syntax highlighted."
    end
    ```

## Skip on Success <r label=skip-on-success/>

Skips to the next heading (any level) if the block executed successfully

    ``` bash skip-on-success
    true
    ```

You will never see this message.

## Skip on Failure <r label=skip-on-failure/>

Skips to the next heading (any level) if the block didn't execute successfully

    ``` bash skip-on-failure
    nonexistant_program
    ```

You will never see this message.

## Capture Environment <r label=env/>

The `env` flag will capture any exported variables and make then available to subsequent scripts.

    ``` bash env
    export BLAH=Hello
    ```

Then later:

    ``` bash stdout
    echo $BLAH
    ```

## With <r label=with/>

Allows you to customise what program executes the script.

    ``` json with='kubectl apply -f'
    { "one": "value" }
    ```

Alternatively

    <r with="kubectl apply -f"/>
    ``` json
    { "one": "value" }
    ```

## Borg <r label=borg/>

The code block process replaces the current rundown process. 

This is handy if your script is going to log the user into an SSH session or something, or you want
to return the status code of the script.

    Now logging you into `someserver`.

    ``` bash borg
    ssh user@someserver
    ```

## Shortcodes <r label=shortcodes/>

* Define a shortcode: `<r label=tag/>`
* Add a shortcode description: `<r desc="blah"/>` or `<r desc>Blah</r>`

See [Full shortcode documentation](./shortcodes.md).

## Save <r label=save/>

Used when you want to demonstrate a file's contents, while also saving it for later use.

    ``` yaml save="config.yml"
    apiVersion: v1
    kind: Service
      metadata:
        name: entrypoint
    ```

Later on, you can use the `basename` as an environment variable to use in scripts. The extension specified in the label's value will be used for the temporary file, so the following should output a filename ending in `.yml`.

    ``` bash stdout
    echo "YAML was saved into $CONFIG"
    ```

### Environment Substitution <r label=sub-env/>

When using `save`, you can also apply the `sub-env` flag. This will perform a substitution of any environment variable with it's value. Be careful though, if you refer to a missing environment variable, the block will fail.

    ``` yaml save:config.yml sub-env
    apiVersion: v1
    kind: Service
      metadata:
        name: "$USER-service"
    ```

# Content Modifiers

Content modifiers allow you to change how regular markdown content is renderered.

## Environment Substitution <r label=sub-env-content/>

Only works with the `<rundown/>` tag format.

When using the `sub-env` flag, all enclosed content will have environment substitution activated.

    ``` bash env
    export BLAH=Hello
    ```

    Why, <r sub-env>$BLAH</r> there!

``` bash env
export BLAH=Hello
```

Why, <r sub-env>$BLAH</r> there! But $BLAH isn't substituted this time.

## Early termination

The `stop-ok` and `stop-fail` flags allow you to control early termination points for your script. 

This can be useful when combined with `skip-on-success` flags, where you might want to continue the flow under an error state to display some helpful messages.

Or when using **shortcodes**, where you don't want to fall through to a child heading.

    ``` bash skip-on-success named
    # Is Homebrew installed?
    brew --version
    ```

    You need to install homebrew. Please go to http://brew.sh.

    <r stop-fail/>

