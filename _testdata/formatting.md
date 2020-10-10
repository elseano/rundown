# This is a heading

This text should **be bold**.

This text should be _italic_.

This text `should be code`.

This is [a link](http://www.google.com).

This text:

* Should be
* Bullet points

And this text

``` ruby reveal norun
def is_just_for_show
  true
end
```

We should also see output from an actual run:

``` bash reveal
echo "Hi"
```

Sequential runs should be placed together:

``` bash
echo "Hi"
```

``` bash
echo "Hi"
```

``` bash
echo "Hi"
```

Sequential runs with a middle reveal should be spaced.

``` bash
echo "Hi"
```

``` bash reveal
echo "Hi"
```

``` bash
echo "Hi"
```


## Additionally

Subheadings will be indented.

We also support ~~strikethrough~~!

* Bullets 
* should be 
* indented correctly
* Really long bullet points will be wrapped, and they will be wrapped at the indentation level of the start of the bullet point text so it looks nice.
  * Indented, and the wrapping should also observe that we're indented on wrap an extra two spaces because this is a sub-bullet.
* And back to normal should work as expected as well given we just deintented the line again.
* :beer:

1. So should
2. Numbered
3. Lists


### Indenting

Goes quite deep.

#### But

At 4 levels we don't do much.

However, thematic breaks ignore heading levels.

----

This should still be indented at level 4.

-----

This is a heading
This text should be bold.

This text should be italic.

This text should be code.

This is http://www.google.com|a link.

This text:

• Should be
• Bullet points

And this text

 ┃ def is_just_for_show
 ┃   true
 ┃ end

We should also see output from an actual run:

 ┃ echo "Hi"

✔ Running (Complete)

Sequential runs should be placed together:

✔ Running (Complete)
✔ Running (Complete)
✔ Running (Complete)

Sequential runs with a middle reveal should be spaced.

✔ Running (Complete)

 ┃ echo "Hi"

✔ Running (Complete)
✔ Running (Complete)

  Additionally
  Subheadings will be indented.

  We also support strikethrough!

  • Bullets
  • should be
  • indented correctly
  • Really long bullet points will be wrapped, and they will be wrapped at the
    indentation level of the start of the bullet point text so it looks nice.
    ◦ Indented, and the wrapping should also observe that we're indented on wrap
      an extra two spaces because this is a sub-bullet.

  • And back to normal should work as expected as well given we just deintented
    the line again.
  • 🍺

  1 So should
  2 Numbered
  3 Lists

    Indenting
    Goes quite deep.

      But
      At 4 levels we don't do much.

      However, thematic breaks ignore heading levels.

  ----------------------------------------------------------------------------  
      

      This should still be indented at level 4.
