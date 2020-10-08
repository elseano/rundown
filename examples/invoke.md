# Invocation

Larger rundown scripts tend to require invoking functions.

In Rundown, you can define a function by adding options to a shortcode.

For example:

<r invoke="make-thing" opt-name="Sean"/>

<r stop-ok/>

Do not see this.

# Handy Scripts

## Show the users name <r func="make-thing"/>

<r opt="name" type="string" required>Given $OPT_NAME is a string</r>

Hello, <r sub-env>$OPT_NAME</r>.

