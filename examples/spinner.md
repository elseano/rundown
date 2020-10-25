# Rundown Spinner Examples

By default, a spinner shows when running code blocks.

``` bash
echo
```

The spinner title can be changed by your script, using either `named`:

``` bash named
# Name is the first comment
```

Or, by using the `spinner:"Title"` form:

``` bash spinner:"Custom title via label value"
echo
```


The spinner moves underneath script output when opting to display it when using `stdout`.

``` bash named stdout
# Another custom name
echo "Output from a script"
```

The spinner will indicate failure when a script fails. Execution continues if you're using `ignore-failure`.

``` bash ignore-failure
failed
```

The spinner can also indicate that subsequent scripts don't need to run, such as when using `skip-on-success` and `skip-on-failure`.

``` bash skip-on-success
echo "Hi"
```

## More information

Check out the [Full list of Modifiers](./mods.md) for more information.
