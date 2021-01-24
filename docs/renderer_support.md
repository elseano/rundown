# Rundown - Renderer Support

Rundown's automation tags are provided as *custom HTML elements*, of the form `<r tag another-tag="value"/>`.

The [CommonMark Specification Section 6.8](https://spec.commonmark.org/0.29/#raw-html) details the handling of Raw HTML, which is that the HTML should be left alone to be rendered by the browser. 

Browsers which encounter the Rundown automation tag will consider it to be an unknown tag, and _generally_ won't render it, although this practice is more defacto rather than standard.

In practice some online platforms have an incomplete CommonMark implementation, despite claiming to be fully compliant. BitBucket is a good example, where this has been an issue since 2015.

In other cases, Very Large Organisations™ may have their own internal code management tools (yes, really) which run obscure or home-grown Markdown renderers (yes, really really).

So deciding on one single automation tag format which works transparently everywhere is quite an undertaking. Instead, Rundown relies on the standards (either de jour or defacto), and hopes for the best in terms of readability.

## Alternative

However, all is not lost. There is an alternative, which may work with naughty platforms.

Rundown also supports a comment style automation tag of them form `<!--~ tag tag:value -->`. This is a version of the [Rundown Only Block](./rundown_only_block.md), however it's only treated as an automation tag if it's all on one line.

## Comparison Table

| Platform | Element Style | Comment Style |
| --- | --- | --- |
| GitHub | OK | OK |
| GitLab | OK | Shown |
| BitBucket | Shown | Shown |

