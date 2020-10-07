# Rundown Demonstration File

![Markdown](https://www.markdownguide.org/assets/images/markdown-guide-og.jpg)

Rundown is a CLI tool which turns markdown files into simple programs and scripts.

It's goal is to keep the resulting file easily readable, while providing powerful additional features to your files.

## Execution

By default, rundown runs fenced code blocks.

``` bash reveal named
# Doing some processing work...
sleep 5
```

It also allows you to show the command's output, and receive input.

``` bash reveal named env stdout
# Asking for some information
read -p "Whats your name? " NAME
export NAME
```

Environment variables which were `export`ed by a script are optionally available in subsequent scripts.

``` bash reveal stdout
echo "Hello, $NAME"
```

## Shortcodes <r label=my-shortcode/>

This heading has a **shortcode**, which you can see next to the heading itself.

In Rundown **shortcodes** enable you to skip directly to that heading, either via additional command line options, or by asking rundown to present a menu.

``` bash named
# Waiting
sleep 5
```

## Formatting

Rundown renders markdown to the console, and it supports all of markdowns rendering codes:

* This is a bullet list.
  * This is a child bullet list
* Back to the original level.

1. Numbered lists are also possible.
2. This is number two.

Rundown will also perform syntax highlighting on scripts it displays to the console. Take this ruby program for example:

``` ruby norun reveal
class Someclass < Object
  def do_a_thing(name)
    puts "Why, hello ${name}"
  end
end
```

What are you waiting for? Give Rundown a try today, and make the terminal pretty again!