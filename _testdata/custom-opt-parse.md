# Options <r label="options"/>

<r opt="name" type="string" desc="Name" required/>

<r sub-env stdout/>

``` bash
echo "The name is $OPT_NAME"
```

# Enum <r label="options:enum"/>

<r opt="thing" type="enum|yep|nah" desc="Yep/Nah" required/>

Result: <r sub-env>$OPT_THING</r>

-----

Options

  Output
  ‣ The name is Blah
  ✔ Running (Complete)