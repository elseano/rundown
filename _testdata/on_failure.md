# OnFailure Handling

Test of the `on-failure` attribute.

## Success Case <r section="success"/>

``` bash
true
```

<r on-failure>You should only see me during a failure</r>

## Fail Case <r section="search"/>

``` bash
failed
```

<r on-failure>You should only see me during a failure.</r>

<r on-failure="command not found">Please install the command.</r>

<r on-failure="no match">You won't see me in rundown.</r>

-----

OnFailure Handling

  Test of the  on-failure  attribute.


  ## Success Case

  ✔ Running (Complete)

  ## Fail Case

  ✖ Running (Failed)


Error - exit status 127 in:

  1: #!/usr/bin/env bash
  2: 
  3: set -Eeuo pipefail
  4: 
  5: failed

SCRIPT: line 5: failed: command not found


You should only see me during a failure.

Please install the command.
