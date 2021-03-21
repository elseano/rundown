# Prompt Manual Tests

These tests prompt the user for input.

## Query and Echo <r label="test1"/>

<r opt="query" prompt="Enter your name" type="string" desc="Name" required />

<r stdout/>

``` bash
echo $OPT_QUERY
```

Done.