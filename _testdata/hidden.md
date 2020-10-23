# Hidden Content

There's two types of hidden content in Rundown.

1. Content thats hidden from markdown, but visible to rundown
2. Content hidden from rundown, but visible to markdown.

Each method employs a different technique due to the nature of markdown.

## Hiding content from markdown

In Rundown hidden blocks are a great way to hide execution detail from readers of markdown, but make them visible to rundown.

    <!--~
    I'll only be rendered inside Rundown.
    -->

Pay attention to the `~` at the end of opening marker, this denotes a rundown hidden block. Without the `~`, it's just a normal markdown comment block.

<!--~
I'll only be rendered inside Rundown.
-->

### What can go inside hidden blocks

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

### Uses for Hidden Blocks

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

~~~ markdown reveal norun
``` bash skip-on-failure
ifail
```

<!--~
## Error
-->
~~~

    Will skip to here.

## Hiding content from Rundown

There are times when you'll want to hide content from being executed or displayed in Rundown, while making it visible in Markdown.

To achieve this, use the `<rundown>` or `<r>` tag, with the `ignore` attribute.

    <r ignore>Content</r>

You can hide just a few words in a paragraph, or ignore multiple paragraphs and code blocks.

If you're ignoring multiple paragraphs (or block level elements), make sure your `<rundown>` tags are on **their own lines**, separated with blank lines:

    <r ignore>

    > This is a blockquote

    </r>



-----

Hidden Content

  There's two types of hidden content in Rundown.
  
  1. Content thats hidden from markdown, but visible to rundown
  2. Content hidden from rundown, but visible to markdown.
  
  Each method employs a different technique due to the nature of markdown.
  

  ## Hiding content from markdown

  In Rundown hidden blocks are a great way to hide execution detail from 
  readers of markdown, but make them visible to rundown.
  
    <!--~
    I'll only be rendered inside Rundown.
    -->
  
  Pay attention to the  ~  at the end of opening marker, this denotes a 
  rundown hidden block. Without the  ~ , it's just a normal markdown comment
  block.
  
  I'll only be rendered inside Rundown.


  ### What can go inside hidden blocks

  You can put whatever you like inside these hidden blocks: headings, code
  blocks, etc. They'll operate as if the comment markers aren't there at all.

    echo "I'm running from inside a rundown hidden block."

  Output
  ‣ I'm running from inside a rundown hidden block.
  ✔ Running (Complete)

  This includes headings with shortcodes, or code setup blocks. Unless your
  Markdown renderer reveals comments, these blocks wont be visible, but 
  Rundown will see them.


  ## Hidden heading 

  This is a hidden heading. You won't see it in markdown.


  ### Uses for Hidden Blocks

  Hidden blocks are a great way to provide "progressive enhancement" to your
  Rundown scripts, such as asking for input when running via rundown, but 
  hiding that code when viewing via markdown.

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

  Hidden headings are good when you want to  skip-on-failure  or  skip-on-success
  without actually creating a new heading: 

    ``` bash skip-on-failure
    ifail
    ```
    
    <!--~
    ## Error
    -->
    
    Will skip to here.


  ## Hiding content from Rundown 

  There are times when you'll want to hide content from being executed or
  displayed in Rundown, while making it visible in Markdown.

  To achieve this, use the  <rundown>  or  <r>  tag, with the  ignore  
  attribute.

    <r ignore>Content</r>

  You can hide just a few words in a paragraph, or ignore multiple paragraphs
  and code blocks.

  If you're ignoring multiple paragraphs (or block level elements), make sure
  your  <rundown>  tags are on their own lines, separated with blank lines:

    <r ignore>
  
    > This is a blockquote
  
    </r>
