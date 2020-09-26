# Rundown RPC functionality

Rundown supports RPC functionality which enables your scripts to alter how rundown displays and handles process execution.

Behind the scenes, rundown uses this to change behaviour, but you also have direct access to the RPC queue.

There's two modes to Rundown's RPC:

1. File based, using the `$RUNDOWN` environment varible which contains the file path.
2. Writing the `\e[R` escape code to STDOUT.

The escape code functionality hasn't been built yet.

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
echo "\e[R Title changed yet again"
```

## Passing environment variables

Code blocks can set environment variables, and they'll be provided to subsequent environment variables.

``` bash reveal nospin
echo "env: SOMEVAL=Test" > $RUNDOWN
echo "\e[R env: MODE=thing"
```

Now we can reference that later.

``` bash stdout reveal nospin
echo $SOMEVAL
echo $MODE
```