# Stdout Tests

Scripts can write to STDOUT. By default, this is hidden.

``` bash nospin
sleep 20
echo "You won't see this"
```

We can reveal STDOUT easily.

``` bash stdout
sleep 1
echo "You should see this"
```

STDOUT will be indented, and correctly formatted when showing progress:

``` bash stdout
printf "Hello\r"
sleep 2
printf "World"
```


STDOUT is also smart enough to hide the spinner when waiting for input on the same line:

``` bash stdout
sleep 1
read -p "Enter something: " something
echo "Got $something"
```

However, STDOUT will reveal the spinner if your input prompt is a blank line. To overcome that, use the `nospin` modifier.

``` bash stdout nospin
read something
```

Ok done.

