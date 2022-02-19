# Serving a Rundown File

Welcome to the DailyCare administration activites.

Using these activities should be done with extreme care, as they're destructive and there's **no undo**!

## Archive a Search <r label="archive"/>

<r desc>Archives a search into an Archive account.</r> In order to use this, we need the search ID.

<r opt="Search ID" type="string" desc="The search ID to archive" />

<r stdout  />

``` bash
echo "1"
sleep 1
echo "2"
sleep 1
echo "3"
sleep 1
echo "4"
```

With that completed, we can run the other bits.

<r stdout/>

``` bash
echo "Hello"
sleep 1
```

And now, finally, the good stuff!

<r stdout/>

``` bash
echo "Hello"
sleep 1
```

### Some nested task <r label="archive:end"/>

This is the end.
