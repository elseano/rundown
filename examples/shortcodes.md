# Rundown Shortcodes

In Rundown, ShortCodes allow you to create subcommands within your markdown file. ShortCodes are special markup inside a heading:

``` markdown reveal norun
# This is a heading <r label=my-shortcode />

You can run me via `rundown shortcodes.md my-shortcode`
```

With rundown autocompletion added to your shell, you can see the shortcodes within a markdown file via tab completion.

You can also view all shortcodes of a markdown file via `rundown c shortcodes.md`. You can provide descriptions for your shortcodes via the rundown element:

``` markdown reveal norun
# This is a heading <r label=my-shortcode/>

<r desc="This is a description for the heading"/>

# This is another heading <r label=my-other-shortcode/>

<r desc>This is a description which is also visible on execution.</r>
```

When headings with shortcodes are rendered, the shortcode will be printed next to the heading.

## I am a heading with a shortcode <r label=my-shortcode />

You should see `(my-shortcode)` next to the heading above when running this file in Rundown, and you should see only the heading in GitHub.

<r desc>Shortcodes can have descriptions</r>

## Shortcode Setup Blocks <r label=setup-blocks/>

Shortcodes interrupt the standard top-down flow of a markdown file. Longer files may refer to environment variables or actions preformed in previous steps, but because you're using shortcodes, these steps have been skipped.

<r desc>You can mark a code block as a `setup` block.</r> These blocks are special, as firstly, will be run prior to any child headings when skipping directly to that child heading, and secondly, they only run once per rundown execution. So if you're running two child headings in the same execution, the setup is only executed by the first shortcode.

~~~ markdown reveal norun
# Deploying

Before deploying, you'll want to make sure you've set your environment correctly.

``` bash named env setup
# Ensure you've set your AWS Profile
export AWS_PROFILE=dev-environment
```

## Deploy Proxies <r label=deploy-proxies/>

``` bash named
# Deploying proxies
kubectl apply -f deploy_proxies.yml
```
~~~

In the above, if we run the `deploy-proxies` shortcode directly, the preceding `setup` block will also be run.

### Recommended usage of setup blocks <r label="recommended"/>

<r desc="Some recommendations on how to use shortcodes"/>
<r opt="flag" type="bool" required default="false" desc="Activate the flag"/>

Structure your setup blocks to ensure that subsequent scripts fail for the right reason. For example, you probably don't want to install anything in a setup script, but you might want to verify something is installed.

## Shortcode arguments

Shortcodes can accept arguments when invoked from the command line. Here's a shortcode which defines 3 arguments:

~~~ html reveal norun
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

And the shortcode would be invoked like this:

``` bash
rundown README.md wait-complete +name="Cool Bro"
```

