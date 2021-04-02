``` bash env nospin
export BLAH=one
```

Hello <r sub-env>**$BLAH**</r>.

And hi to <r sub-env>$MISSING</r>.

Subenv also works with the block form of rundown:

<r sub-env>$BLAH</r>

And we can sub-env inside code blocks:

<r norun reveal sub-env/>

``` http
GET http://localhost:3000/$BLAH HTTP/1.1
```

Done. 

-----

Hello one.

  And hi to $MISSING (not set).

  Subenv also works with the block form of rundown:

  one

  And we can sub-env inside code blocks:

    GET http://localhost:3000/$BLAH HTTP/1.1
  
  Done.