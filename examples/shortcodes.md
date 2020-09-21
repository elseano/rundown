# Rundown Shortcodes

In Rundown, ShortCodes allow you to create subcommands within your markdown file. ShortCodes are special markup inside a heading:

``` markdown reveal norun
# This is a heading [](label:my-shortcode)

You can run me via `rundown shortcodes.md my-shortcode`
```

With rundown autocompletion added to your shell, you can see the shortcodes within a markdown file via tab completion.

You can also view all shortcodes of a markdown file via `rundown c shortcodes.md`.

When headings with shortcodes are rendered, the shortcode will be printed next to the heading.

## I am a heading with a shortcode [](label:shortcode)

You should see (shortcode) next to the heading above.

## Shortcode Prerequisites

Shortcodes interrupt the standard top-down flow of a markdown file. Longer files may refer to environment variables or actions preformed in previous steps, but because you're using shortcodes, these steps have been skipped.

You can mark a code block as a `setup` block. These blocks are special, as firstly, the only run once per rundown execution. They also will be run prior to any child headings when skipping directly to them.

    # Deploying

    Before deploying, you'll want to make sure you've set your environment correctly.

    ``` bash named env setup
    # Ensure you've set your AWS Profile
    export AWS_PROFILE=dev-environment
    ```

    ## Deploy Proxies [](label:deploy-proxies)

    ``` bash named
    # Deploying proxies
    kubectl apply -f deploy_proxies.yml
    ```

In the above, if we run the `deploy-proxies` shortcode directly, the preceding `setup` block will also be run.

### Recommended usage of setup blocks

Structure your setup blocks to ensure that subsequent scripts fail for the right reason. For example, you probably don't want to install anything in a setup script, but you might want to verify something is installed.