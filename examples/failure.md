# Handling Script Failure

Rundown has a few different ways it handles and interprets errors.

## Errors with skip-on-success

If your script errors with the `skip-on-success` modifier, the error is swallowed and flow continues.

``` bash skip-on-success
idontexistcall
```

If your script succeeds with `skip-on-success`, then flow skips to the next heading:

``` bash skip-on-success
true
```

You won't see this.

#### Target

Flow will skip to here.

## Errors with skip-on-failure

If `skip-on-failure` succeeds, then flow continues.

``` bash skip-on-failure
true
```

If the script errors, then flow jumps to the next heading.

``` bash skip-on-failure
idontexistcall
```

You shouldn't see this.

#### Target

Flow will skip to here.

## Supporting process exit

Scripts can exit with a non-zero code which will be treated as a failure.

``` bash skip-on-failure
echo "This is a failure"
exit 1
```


## Errors in normal scripts

Without the `skip-on-failure` or `ignore-failure` flags, a failing script will terminate the rundown file.

You can add helpful troubleshooting documentation for failure cases using the `<rundown>` tag with the `on-failure` attribute. The contents of the tag will be shown prior to exiting.

The tag supports two modes:

* `<r on-failure>` Will show for any failure.
* `<r on-failure="regexp">` Will show if the stdout/stderr stream matches the given regexp.

The `on-failure` tags apply to the current heading only. Multiple tags are supported, and will be shown in the order they are declared.

For example:

~~~ markdown reveal norun
# Heading

``` bash
somemissingapp
```

<r on-failure>:dazed: There was a failure.</r>
<r on-failure="not found">You need to install acme</r>
~~~

This will render both tags, as `somemissingapp` will result in a `Command not found` error printed out. You can hide these tags inside a hidden block if you don't want to reveal the error conditions to readers.

![On Failure Screenshot](../_images/on-failure.png)

## Raw failures

A raw failure looks like this:

``` bash
idontexistcall
```