package rundown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdoutFormatting(t *testing.T) {
	code := `STDOUT will be indented, and correctly formatted when showing progress:

<r stdout/>

~~~ bash
printf "Hello\r"
printf "World"
~~~

STDOUT is also smart enough to hide the spinner when waiting for input on the same line:`

	expected := `STDOUT will be indented, and correctly formatted when showing progress:

Running...
  Hello\rWorld
Running (.*)

STDOUT is also smart enough to hide the spinner when waiting for input on the same line:`

	buffer := bytes.Buffer{}

	loaded, _ := LoadString(code, "test.md")
	loaded.MasterDocument.Render(&buffer)

	assert.Regexp(t, expected, buffer.String())
}
