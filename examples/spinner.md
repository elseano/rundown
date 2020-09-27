# Rundown Spinner Examples

By default, a spinner shows when running code blocks.

``` bash
echo
```

The spinner title can be changed by your script.

``` bash named
# Custom name
```

The spinner move underneath script output when opting to display it.

``` bash named stdout
# Another custom name
echo "Output from a script"
```

The spinner will indicate failure when a script fails.

``` bash ignore-failure
failed
```

The spinner can also indicate that subsequent scripts don't need to run.

``` bash skip-on-success
echo "Hi"
```