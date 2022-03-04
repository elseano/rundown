<r opt="opt-one" type="string" required desc="Activate the thing"/>

# Help using this file <r label="rundown:help"/>

To use this file, make sure you specify the `opt-one` parameter. For example:

```
rundown go --opt-one=Thing
```

# This is a test for Options <r section="go" />

Hi there <r sub-env>$OPT_OPT_ONE</r>.

-----

Help using this file

  To use this file, make sure you specify the  +opt-one  parameter.

  
Error: '+opt-one' is required