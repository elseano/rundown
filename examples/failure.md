# Failing scripts

Rundown has a few different ways it handles errors.

## Errors with skip_on_success

If your script errors with the `skip_on_success` modifier, the error is swallowed and flow continues.

``` bash skip_on_success
idontexistcall
```

If your script succeeds with `skip_on_success`, then flow skips to the next heading.

``` bash skip_on_success
true
```

You won't see this.

## Errors with skip_on_failure

If `skip_on_failure` succeeds, then flow continues.

``` bash skip_on_failure
true
```

If the script errors, then flow jumps to the next heading.

``` bash skip_on_failure
idontexistcall
```

You shouldn't see this.

## Errors in normal scripts

This demonstrates the output from a failing script.

``` bash
idontexistcall
```