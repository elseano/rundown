# This is a heading

Some `body *match*` text *italic* and **bold**, and ~~nope~~.

# Code Block Parsing

[](named)

``` ruby
# One
def somecode(); end
```

<!--~ named -->

``` ruby
# Two
def othercode(); end
```

``` ruby named
# Three
def method
  i = 5
end
```

That was _some_ *code* **alright**!

``` go reveal norun
func (r *Something) doThing() {
  fmt.Println("This is a go output")
}
```

This is a list:

* One
* Two
* Three

This is a numbered list:

1. One
2. Two