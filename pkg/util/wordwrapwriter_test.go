package util

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapBasic(t *testing.T) {
	var buf bytes.Buffer

	w := NewWordWrapWriter(&buf, 10)
	w.Write([]byte("This is some text which has a really long line\nHere's another line."))

	assert.Equal(t, "This is\nsome text\nwhich has\na really\nlong line\nHere's\nanother\nline.", buf.String())
}

func TestWrapCR(t *testing.T) {
	var buf bytes.Buffer

	w := NewWordWrapWriter(&buf, 10)
	w.Write([]byte("This is some text which has a really long line\nHere's\ranother line."))

	assert.Equal(t, "This is\nsome text\nwhich has\na really\nlong line\nHere's\ranother\nline.", buf.String())
}

func TestWrapCREnd(t *testing.T) {
	var buf bytes.Buffer

	w := NewWordWrapWriter(&buf, 10)
	w.Write([]byte("This is some text which has a really long line\nHere's\ranother line.\r"))

	assert.Equal(t, "This is\nsome text\nwhich has\na really\nlong line\nHere's\ranother\nline.\r", buf.String())
}

func TestWrapCallback(t *testing.T) {
	var buf bytes.Buffer

	w := NewWordWrapWriter(&buf, 10)
	w.SetAfterWrap(func(out io.Writer) int {
		n, _ := out.Write([]byte(" "))
		return n
	})
	w.Write([]byte("This is some text which has a really long line\nHere's\ranother line.\r"))

	assert.Equal(t, "This is\n some text\n which has\n a really\n long line\nHere's\r another\n line.\r ", buf.String())
}
