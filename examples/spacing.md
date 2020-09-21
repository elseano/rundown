Sequential runs should appear together:

``` bash
echo ""
```

``` bash
echo ""
```

Text after code blocks should be separated by a blank line.

This should also work when the script doesn't print out a newline at the end.

``` bash stdout
printf "Ending on same line"
```

This should have a similar gap.

Multiple output scripts together should be together.

``` bash stdout
printf "Ending on same line"
```

``` bash stdout
printf "Ending on same line"
```

This should have a similar gap.