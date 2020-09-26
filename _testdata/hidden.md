# Hidden Blocks and how they work

In Rundown, you can add hidden blocks are a great way to hide execution detail from readers, but add enhancements when running the markdown file using rundown.

    <!--~
    I'll only be rendered inside Rundown.
    -->

Pay attention to the `~` at the end of opening marker, this denotes a rundown hidden block. Without the `~`, it's just a normal markdown comment block.

<!--~
I'll only be rendered inside Rundown.
-->

You can put whatever you like inside these hidden blocks: headings, code blocks, etc. They'll operate as if the comment markers aren't there at all.

<!--~
``` bash reveal
echo "I'm running from inside a rundown hidden block."
```
-->

Headings within a hidden block aren't included in `skip-on-success` and `skip-on-failure` flags, so be careful of that.

-----

Hidden Blocks and how they work
In Rundown, you can add hidden blocks are a great way to hide execution detail
from readers, but add enhancements when running the markdown file using rundown.

 ┃ <!--~
 ┃ I'll only be rendered inside Rundown.
 ┃ -->
 
Pay attention to the ~ at the end of opening marker, this denotes a rundown
hidden block. Without the ~, it's just a normal markdown comment block.

I'll only be rendered inside Rundown.

You can put whatever you like inside these hidden blocks: headings, code
blocks, etc. They'll operate as if the comment markers aren't there at all.

 ┃ echo "I'm running from inside a rundown hidden block."

✔ Running (Complete)

Headings within a hidden block aren't included in skip-on-success and
skip-on-failure flags, so be careful of that.
