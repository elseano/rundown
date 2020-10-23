
# Rundown RPC functionality

Rundown supports RPC functionality.

This enables your scripts to alter how rundown displays and handles process
execution.

## Changing the spinner title

Rundown RPC supports setting the spinner title via:

    echo "name: New spinner title" > $RUNDOWN

For example:

``` bash
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

Code blocks can set environment variables, and they'll be provided to
subsequent environment variables.

``` bash reveal
echo "env: SOMEVAL=Test" > $RUNDOWN
```

Now we can reference that later.

``` bash stdout reveal nospin
echo $SOMEVAL
```

There's also a shorthand to capture **exported** environment variables from a
script, using the `env` modifier.

``` bash reveal env
export NEW_VAR=Something
```

And then reference again:

``` bash reveal stdout
echo $NEW_VAR
```


-----

Rundown RPC functionality

  Rundown supports RPC functionality.

  This enables your scripts to alter how rundown displays and handles process
  execution.


  ## Changing the spinner title

  Rundown RPC supports setting the spinner title via:

    echo "name: New spinner title" > $RUNDOWN

  For example:

  ✔ Title changed again (Complete)

  ## Passing environment variables

  Code blocks can set environment variables, and they'll be provided to
  subsequent environment variables.

    echo "env: SOMEVAL=Test" > $RUNDOWN

  ✔ Running (Complete)

  Now we can reference that later.

    echo $SOMEVAL

  Output
  ‣ Test
  
  There's also a shorthand to capture exported environment variables from a
  script, using the  env  modifier.

    export NEW_VAR=Something

  ✔ Running (Complete)

  And then reference again:

    echo $NEW_VAR

  Output
  ‣ Something
  ✔ Running (Complete)
