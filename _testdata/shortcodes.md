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