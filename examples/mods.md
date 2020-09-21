# Modifier formats

Modifiers come in two flavours: Flags, and Values. Flag modifiers are single `words`, which simply imply they're activated. Values are a `key:value` format, and the value can optionally be quoted with either single or double quotes if it contains spaces.

Modifiers can be specified using one of three possible formats. Markdown renderers tend to vary, and in order to keep a document readable, you should choose the right format.

## Fenced code block format

Modifiers on fenced code blocks are generally the preferred way of altering execution. This format works well with GitHub rendering.

    ``` bash nospin
    sleep 1
    ```

However, there's situations where you need to modify something which isn't a fenced code block, or your render breaks when having multiple words on the syntax line.

## Invisible Link format

This format is a markdown inline element, which means you can place it inside block elements (headings, paragraphs, etc). It will render HTML, however HTML links with no text aren't rendered by browsers, so these will be invisible to readers.

    [](nospin)
    ``` bash
      sleep 1
    ```

[](nospin)
``` bash
  sleep 1
```

## Comment format

Finally there's the comment format. This will render to HTML as well, but as it's a comment, browsers won't display it. Some client-side Javascript markdown renderers do still display this.

    <!--~ nospin -->
    ``` bash
    sleep 1
    ```

<!--~ nospin -->
``` bash
sleep 1
```

Take care to ensure your comment formats remain as a single line entity, and include the tilde `~` after the comment starting marker. 

Multiple line comments with the tilde (`~`) are Invisible Blocks, which serve a different purpose in Rundown. You can find out more about them at the [Invisible Blocks](./invisible.md) page.

# Testing Modifiers

Modifiers allow you to change how the code executes.

## Blank

Default execution

``` bash
sleep 1
```

## nospin

This prevents the spinner from showing

``` bash nospin
sleep 1
```

## named

The first line of the script is treated as a comment, and that comment body is the title of the spinner.

``` bash named
# Waiting...

sleep 1
```

## Stdout

This shows STDOUT in the output.

``` bash
echo "You won't see this"
```

``` bash stdout
echo "You should see this"
```

## norun

This ignores the code block, don't run it, and don't display it.

``` bash norun
echo "This will be displayed and not run"
```

## reveal

This will display the code block **AND** run it.

``` bash reveal
sleep 1
```

A common approach to rendering syntax highlighted code in the console is to use `norun` with `reveal`:

    ``` ruby norun reveal
    def method()
      "Hi, I'm syntax highlighted."
    end
    ```

## Skip on Success

Skips to the next heading (any level) if the block executed successfully

``` bash skip_on_success
true
```

You will never see this message.

## Skip on Failure

Skips to the next heading (any level) if the block didn't execute successfully

``` bash skip_on_failure
nonexistant_program
```

You will never see this message.

## Capture Environment

The `env` flag will capture any exported variables and make then available to subsequent scripts.

    ``` bash env
    export BLAH=Hello
    ```

Then later:

    ``` bash stdout
    echo $BLAH
    ```

## With

Allows you to customise what program executes the script.

    ``` json with:'kubectl apply -f'
    { "one": "value" }
    ```

## Save

Used when you want to demonstrate a file's contents, while also saving it for later use.

    ``` yaml save:config.yml
    apiVersion: v1
    kind: Service
      metadata:
        name: entrypoint
    ```

Later on, you can use the `basename` as an environment variable to use in scripts. The extension specified in the label's value will be used for the temporary file, so the following should output a filename ending in `.yml`.

    ``` bash stdout
    echo "YAML was saved into $CONFIG"
    ```

### Environment Substitution

When using `save`, you can also apply the `env_aware` flag. This will perform a substitution of any environment variable with it's value. Be careful though, if you refer to a missing environment variable, the block will fail.

    ``` yaml save:config.yml env_aware
    apiVersion: v1
    kind: Service
      metadata:
        name: "$USER-service"
    ```
