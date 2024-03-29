# Stdout Tests <r section="stdout"/>

Scripts can write to STDOUT. By default, this is hidden.

``` bash
echo "You won't see this"
```

We can reveal STDOUT easily.

<r stdout/>

``` bash
echo "You should see this"
```

STDOUT will be indented, and correctly formatted when showing progress:

<r stdout/>

``` bash
printf "Hello\r"
printf "World"
```

STDOUT is also smart enough to hide the spinner when waiting for input on the same line:

<r stdout/>

``` bash
read -t 5 -p "Enter something: " something || true
```

However, STDOUT will reveal the spinner if your input prompt is a blank line.

To overcome that, use the `nospin` modifier.

<r stdout nospin/>

``` bash
read -t 5 -p "Enter something:
" something || true
```

If you're running a `named` code segment, the output heading will be the name.


<r stdout named/>

``` bash 
# This is the title
echo "Hi there"
```

Ok done.

-----

Stdout Tests

  Scripts can write to STDOUT. By default, this is hidden.
  
  ✔ Running (Complete)
  
  We can reveal STDOUT easily.
  
  Output
  ‣ You should see this
  ✔ Running (Complete)
  
  STDOUT will be indented, and correctly formatted when showing progress:
  
  Output
  ‣ World
  ✔ Running (Complete)
  
  STDOUT is also smart enough to hide the spinner when waiting for input on 
  the same line:
  
  Output
  ‣ Enter something: 
  ✔ Running (Complete)
  
  However, STDOUT will reveal the spinner if your input prompt is a blank 
  line.
  
  To overcome that, use the  nospin  modifier.
  
  Output
  ‣ Some prompt
  
  If you're running a  named  code segment, the output heading will be the 
  name.
  
  This is the title
  ‣ Hi there
  ✔ This is the title (Complete)
  
  Ok done.
