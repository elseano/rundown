# Testing various scripts which ask for input

Scripts can ask for input and display their output.

``` bash stdout
echo "Hi"
```

You can also run more complex commands which ask a series of inputs.

``` bash stdout env
read -p "Hi " OUTPUT
```

Again:

``` bash stdout env
echo "Last input was $OUTPUT"
read -p "Hi " OUTPUT
```