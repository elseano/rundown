# Functions

Rundown supports the idea of functions within your markdown file. Functions are defined similarly to ShortCodes, by using headings.

## Defining Functions

Defining a function does not change the default execution order of your Rundown file - that is, functions will be executed as they're encountered, as well as when they're explicitly invoked.

    # Some Function <r func="some:func"/>

The function is defined as anything between the definition heading and the next heading of the same level number or less.

### This is a function <r func="some:func"/>

Anything can be inside a function - text, fenced code blocks, etc. Text will be displayed, fenced code blocks will be executed, as normal.

<r spinner="Things"/>

``` bash
true
```

## Invoking functions

To invoke a function, use the invoke attribute.

    <r invoke="some:func"/>

The function contents will be inserted at the point of the invoke call, excluding the function's initial heading.

<r invoke="some:func"/>

## Function Parameters

Functions support parameters via the `opt` attribute (the same attribute used for document and shortcode parameters).

Just like ShortCode parameters, the parameter must be under the function's heading:

    # Some Function <r func="some:func"/>

    <r opt="force" desc="Force override" default="true" required type="boolean"/>

When invoking the function, you can supply the parameters as attributes. To prevent any conflict, the parameters must be prefixed with `opt-`:

    <r invoke="some:func" opt-force="true"/>

## Functions for Readers

The function invoke tag can contain text within it. This text can be used for document readers, as Rundown will not render the provided contents:

    <r invoke="some:func">Use the handy script from above to make sure it worked</r>

<r stop-ok/>
-----

Functions

  Rundown supports the idea of functions within your markdown file. Functions
  are defined similarly to ShortCodes, by using headings.


  ## Defining Functions

  Defining a function does not change the default execution order of your
  Rundown file - that is, functions will be executed as they're encountered, as
  well as when they're explicitly invoked.

    # Some Function <r func="some:func"/>

  The function is defined as anything between the definition heading and the
  next heading of the same level number or less.


  ### This is a function

  Anything can be inside a function - text, fenced code blocks, etc. Text will
  be displayed, fenced code blocks will be executed, as normal.

  ✔ Things (Complete)

  ## Invoking functions

  To invoke a function, use the invoke attribute.

    <r invoke="some:func"/>

  The function contents will be inserted at the point of the invoke call,
  excluding the function's initial heading.

  ✔ Things (Complete)

  ## Function Parameters

  Functions support parameters via the  opt  attribute (the same attribute
  used for document and shortcode parameters).

  Just like ShortCode parameters, the parameter must be under the function's
  heading:

    # Some Function <r func="some:func"/>

    <r opt="force" desc="Force override" default="true" required
  type="boolean"/>

  When invoking the function, you can supply the parameters as attributes. To
  prevent any conflict, the parameters must be prefixed with  opt- :

    <r invoke="some:func" opt-force="true"/>


  ## Functions for Readers

  The function invoke tag can contain text within it. This text can be used
  for document readers, as Rundown will not render the provided contents:

    <r invoke="some:func">Use the handy script from above to make sure it
  worked</r>

