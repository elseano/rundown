# Testing various scripts which ask for input

Given scripts can display their output:

``` bash stdout
echo "Hi"
```

Then you're able to run more complex commands which ask for inputs. 

``` bash stdout env
read -p "Hi " OUTPUT
export OUTPUT
```

Note that this only works with `stdout` applied. If a scirpt asks for input, and doesn't have `stdout` applied, then your rundown file will hang waiting for input without the users knowledge.

Again:

``` bash stdout env
echo "Last input was $OUTPUT"
read -p "Hi " OUTPUT
```

For complex interactions, such as launching an interactive subshell, it's recommended you allow that process to completely take over rundown via the `borg` flag. Use this mode sparingly as it doesn't allow your rundown file to continue afterwards.

Type exit to return to your current shell.

``` bash borg
sh
```