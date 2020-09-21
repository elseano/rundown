# Rundown RPC functionality

Rundown supports RPC functionality which enables your scripts to alter how rundown displays and handles process execution.

## Changing the spinner title

Rundown RPC supports setting the spinner title via:

    echo "name: New spinner title" > $RUNDOWN

For example:

``` bash stdout
echo "I should appear on STDOUT"
echo "Some other text"
echo "Some\nMultiline\nText"
echo "Rundown is at ${RUNDOWN}"
echo "name: Title was changed" > $RUNDOWN
# sleep 2
echo "name: Title changed again" > $RUNDOWN
# sleep 2
```

## Passing environment variables

Code blocks can set environment variables, and they'll be provided to subsequent environment variables.

``` bash reveal nospin
echo "env: SOMEVAL=Test" > $RUNDOWN
```

Now we can reference that later.

``` bash stdout reveal nospin
echo $SOMEVAL
```