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

## Shortcode Setup Blocks

Shortcodes interrupt the standard top-down flow of a markdown file. Longer files may refer to environment variables or actions preformed in previous steps, but because you're using shortcodes, these steps have been skipped.

You can mark a code block as a `setup` block. These blocks are special, as firstly, will be run prior to any child headings when skipping directly to that child heading, and secondly, they only run once per rundown execution. So if you're running two child headings in the same execution, the setup is only executed by the first shortcode.

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

In the above, if we run the `deploy-proxies` shortcode directly, the preceding `setup` block will also be run.

### Recommended usage of setup blocks

Structure your setup blocks to ensure that subsequent scripts fail for the right reason. For example, you probably don't want to install anything in a setup script, but you might want to verify something is installed.