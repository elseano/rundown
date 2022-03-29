package term

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColourOutput(t *testing.T) {
	input := []byte("total 40\ndrwxr-xr-x   7 elseano  staff   224 21 Feb 15:00 \033[34m.\033[39;49m\033[0m\ndrwxr-xr-x  36 elseano  staff  1152 23 Mar 17:39 \033[34m..\033[39;49m\033[0m\n-rw-r--r--   1 elseano  staff    91 18 Feb 14:35 Dockerfile.alpine\n-rw-r--r--   1 elseano  staff    67 21 Feb 15:06 Dockerfile.exec.busybox\n-rw-r--r--   1 elseano  staff   295 21 Feb 17:15 Dockerfile.ubuntu\n-rwxr-xr-x@  1 elseano  staff   560 16 Sep  2020 \033[31mbash_autocomplete\033[39;49m\033[0m\n-rwxr-xr-x   1 elseano  staff  1755 10 Oct  2020 \033[31minstall\033[39;49m\033[0m\n")
	reader := bytes.NewBuffer(input)

	buffer := NewStdoutBuffer()
	buffer.Process(reader)
	output, lines := buffer.String()

	if !assert.Equal(t, 9, lines) {
		t.Logf("%#v", output)
	}
}

func TestIndicatorOutput(t *testing.T) {
	input := []byte("1%\r50%\r100%\n")
	reader := bytes.NewBuffer(input)

	buffer := NewStdoutBuffer()
	buffer.Process(reader)
	output, lines := buffer.String()

	if !assert.Equal(t, 2, lines) {
		t.Logf("%#v", output)
	}
}

func TestIndicatorOutput2(t *testing.T) {
	input := []byte("1%\0332K50%\0332K100%\n")
	reader := bytes.NewBuffer(input)

	buffer := NewStdoutBuffer()
	buffer.Process(reader)
	output, lines := buffer.String()

	if !assert.Equal(t, 2, lines) {
		t.Logf("%#v", output)
	}
}
