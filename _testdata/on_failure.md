# OnFailure Handling

Test of the `on-failure` attribute.

## Success Case

``` bash
true
```

<r on-failure>You should only see me during a failure</r>

## Fail Case

``` bash
failed
```

<r on-failure>You should only see me during a failure</r>

-----

OnFailure Handling
Test of the on-failure attribute.

  Success Case
  ✔ Running (Complete)

  Fail Case
  ✖ Running (Failed)

  You should only see me during a failure


exit status 127

✖ Aborted due to failure.
