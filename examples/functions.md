# Rundown Functions

In Rundown, Functions allow you to pull in functionality from elsewhere in your Rundown file, or another Rundown file entirely.

Here's a simple "Hello world" function defined in Rundown:

``` markdown reveal norun
# Hello World <r func=hi />

Hello world.
```

A function ends at either the next heading of the same level, or the end of the file.

The function can then be invoked from within the same file:

``` markdown reveal norun
<r invoke="hi" />
```

Any content you provide to the Rundown tag will be hidden during Rundown execution, but revealed to readers.

``` markdown reveal norun
<r invoke="hi">Now check out the [Hello World](#hello-world) example.</r>
```

You can also invoke functions in other files:

``` markdown reveal norun
<r invoke="hi" from="UTILS.md">Check out the [Hi Code](./UTILS.md#hello-world).</r>
```

## Defining options for your Functions

Options can be defined after your Function heading:

``` markdown reveal norun
<r opt="flag" type="bool" required default="false" desc="Activate the flag"/>
```

Option values are avaiable within your function body as environment varibles prefixed with `OPT_`. For example, the above would be `$OPT_FLAG`.

Any option which has a default, and is required, but not provided will be pre-populated with it's default value. Otherwise, to specify the option value when invoking the function:

``` markdown reveal norun
<r invoke="my-func" opt-flag="true"/>
```

## Treatment of Headings

Because functions content is pulled into your document at the point where it's invoked, headings are adjusted:

* The function heading is not shown at all.
* Any sub-headings within your function are adjusted to be relative to the `invoke` call site's heading.

For example:

~~~ markdown reveal norun
## Heading

<r invoke="func"/>

# Func <r func="func"/>

This is some content.

## Child heading
~~~

Would render as if the content was:

~~~ markdown reveal norun
## Heading

This is some content

### Child heading
~~~

## Documentation

Functions are not included as part of the rundown help output.

# Examples

<r invoke="one" />

<r stop-ok/>

## Function 1 <r func="one"/>

Hi, I'm some content.

### I'm a subheading

I'm sub content.