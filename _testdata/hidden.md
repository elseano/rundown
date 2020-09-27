# Hidden Blocks and how they work

In Rundown, you can add hidden blocks are a great way to hide execution detail from readers, but add enhancements when running the markdown file using rundown.

    <!--~
    I'll only be rendered inside Rundown.
    -->

Pay attention to the `~` at the end of opening marker, this denotes a rundown hidden block. Without the `~`, it's just a normal markdown comment block.

<!--~
I'll only be rendered inside Rundown.
-->

## What can go inside hidden blocks

You can put whatever you like inside these hidden blocks: headings, code blocks, etc. They'll operate as if the comment markers aren't there at all.

<!--~
``` bash reveal stdout
echo "I'm running from inside a rundown hidden block."
```
-->

This includes headings with shortcodes, or code setup blocks. Unless your Markdown renderer reveals comments, these blocks wont be visible, but Rundown will see them.

<!--~
## Hidden heading <r label=hidden>

This is a hidden heading. You won't see it in markdown.
-->

## Uses for Hidden Blocks

Hidden blocks are a great way to provide "progressive enhancement" to your Rundown scripts, such as asking for input when running via rundown, but hiding that code when viewing via markdown.

    <!--~ 
    ``` bash env
    if [ -z ${AWS_REGION:-} ]; then
      read -p "Enter the AWS Region: " AWS_REGION
    fi
    ```
    -->

    Now showing your all EC2 instances in $AWS_REGION:

    ``` bash stdout
    aws ec2 describe-instances
    ```

Hidden headings are good when you want to `skip-on-failure` or `skip-on-success` without actually creating a new heading:

    ``` bash skip-on-failure
    ifail
    ```

    <!--~
    ## Error
    -->

    Will skip to here.


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

  What can go inside hidden blocks
  You can put whatever you like inside these hidden blocks: headings, code
  blocks, etc. They'll operate as if the comment markers aren't there at all.

   ┃ echo "I'm running from inside a rundown hidden block."

  Output
  ‣ I'm running from inside a rundown hidden block.
  ✔ Running (Complete)

  This includes headings with shortcodes, or code setup blocks. Unless your
  Markdown renderer reveals comments, these blocks wont be visible, but Rundown
  will see them.

  Hidden heading 
  This is a hidden heading. You won't see it in markdown.

  Uses for Hidden Blocks
  Hidden blocks are a great way to provide "progressive enhancement" to your
  Rundown scripts, such as asking for input when running via rundown, but hiding
  that code when viewing via markdown.

   ┃ <!--~ 
   ┃ ``` bash env
   ┃ if [ -z ${AWS_REGION:-} ]; then
   ┃   read -p "Enter the AWS Region: " AWS_REGION
   ┃ fi
   ┃ ```
   ┃ -->
   ┃ 
   ┃ Now showing your all EC2 instances in $AWS_REGION:
   ┃ 
   ┃ ``` bash stdout
   ┃ aws ec2 describe-instances
   ┃ ```

  Hidden headings are good when you want to skip-on-failure or skip-on-success
  without actually creating a new heading:

   ┃ ``` bash skip-on-failure
   ┃ ifail
   ┃ ```
   ┃ 
   ┃ <!--~
   ┃ ## Error
   ┃ -->
   ┃ 
   ┃ Will skip to here.
