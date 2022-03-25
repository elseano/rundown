package rundown

import (
	"bytes"
	"testing"

	"github.com/elseano/rundown/testutil"
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

func TestStdoutWithCurl(t *testing.T) {
	code := `Curl seems to write to pty directly
	
<r stdout/>

~~~ bash
curl http://example.org
~~~

`

	expected := `Curl seems to write to pty directly

	<!doctype html>
	<html>
	<head>
			<title>Example Domain</title>

			<meta charset="utf-8" />
			<meta http-equiv="Content-type" content="text/html; charset=utf-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1" />
			<style type="text/css">
			body {
					background-color: #f0f0f2;
					margin: 0;
					padding: 0;
					font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
					
			}
			div {
					width: 600px;
					margin: 5em auto;
					padding: 2em;
					background-color: #fdfdff;
					border-radius: 0.5em;
					box-shadow: 2px 3px 7px 2px rgba(0,0,0,0.02);
			}
			a:link, a:visited {
					color: #38488f;
					text-decoration: none;
			}
			@media (max-width: 700px) {
					div {
							margin: 0 auto;
							width: auto;
					}
			}
			</style>    
	</head>

	<body>
	<div>
			<h1>Example Domain</h1>
			<p>This domain is for use in illustrative examples in documents. You may use this
			domain in literature without prior coordination or asking for permission.</p>
			<p><a href="https://www.iana.org/domains/example">More information...</a></p>
	</div>
	</body>
	</html>`

	buffer := bytes.Buffer{}

	loaded, _ := LoadString(code, "test.md")
	loaded.MasterDocument.Render(&buffer)

	testutil.AssertLines(t, expected, buffer.String())
}

func TestStdoutWithLs(t *testing.T) {
	code := `Curl seems to write to pty directly
	
<r stdout/>

~~~ bash
ls -la --color=always ../build
~~~

`

	expected := `Curl seems to write to pty directly
	
	
	
	`

	buffer := bytes.Buffer{}

	loaded, _ := LoadString(code, "test.md")
	loaded.MasterDocument.Render(&buffer)

	testutil.AssertLines(t, expected, buffer.String())
}
