# Rundown Modifiers

Modifiers in Fenced Code Blocks can be specified either inline on the fence itself, or just prior using a `<rundown/>` tag.

## Inline

``` bash named
# This is a name
true
```

## Tagged

<r named/>

``` bash
# This is a name as well
true
```

## Precedence

By default, if a fenced code block specifies no modifiers, it will attempt to use a preceding `<rundown/>` tag. This can cause issues if your intent was to have **no modifiers** on the code block with an immediately preceding tag.

To solve this, insert an empty tag just prior.

<r desc stdout>This is a heading description, but we don't want it's attributes applied to the next code block.</r>

<r/>

``` bash
echo "You shouldn't see me"
```

-----

Rundown Modifiers
Modifiers in Fenced Code Blocks can be specified either inline on the fence itself,
or just prior using a <rundown/> tag.

  Inline
  ✔ This is a name (Complete)

  Tagged
  ✔ This is a name as well (Complete)

  Precedence
  By default, if a fenced code block specifies no modifiers, it will attempt to use a
  preceding <rundown/> tag. This can cause issues if your intent was to have no
  modifiers on the code block with an immediately preceding tag.

  To solve this, insert an empty tag just prior.

  This is a heading description, but we don't want it's attributes applied to the
  next code block.

  ✔ Running (Complete)
