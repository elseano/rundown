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

## Supporting process exit

Scripts can exit with a non-zero code which will be treated as a failure.

``` bash skip_on_failure
echo "This is a failure"
exit 1
```

You shouldn't see this.

## Errors in normal scripts

This demonstrates the output from a failing script.

``` bash
idontexistcall
```

-----

Failing scripts
Rundown has a few different ways it handles errors.

  Errors with skip_on_success
  If your script errors with the skip_on_success modifier, the error is swallowed and flow continues.

  ✔ Running (Required)

  If your script succeeds with skip_on_success, then flow skips to the next heading.

  ≡ Running (Passed)

  Errors with skip_on_failure
  If skip_on_failure succeeds, then flow continues.

  ✔ Running (Required)

  If the script errors, then flow jumps to the next heading.

  ≡ Running (Passed)

  Supporting process exit
  Scripts can exit with a non-zero code which will be treated as a failure.

  ≡ Running (Passed)

  Errors in normal scripts
  This demonstrates the output from a failing script.

  ✖ Running (Failed)


Error executing script:

#!/usr/bin/env bash

set -Eeuo pipefail

idontexistcall


Error: exit status 127

SCRIPT: line 5: idontexistcall: command not found

✖ Aborted due to failure.
