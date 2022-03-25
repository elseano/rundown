package term

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnsiWritesBasicPrintable(t *testing.T) {
	input := bytes.Buffer{}
	input.WriteString("Hi there")

	output := bytes.Buffer{}

	w := NewAnsiScreenWriter(&input, &output)
	w.Process()

	assert.Equal(t, "Hi there", output.String())
}

func TestAnsiFlushCallbacks(t *testing.T) {
	input := bytes.Buffer{}
	input.WriteString("Hi there")

	output := bytes.Buffer{}

	w := NewAnsiScreenWriter(&input, &output)
	w.BeforeFlush(func() {
		output.WriteString("B")
	})
	w.AfterFlush(func() {
		output.WriteString("A")
	})

	w.Process()

	assert.Equal(t, "BHi thereA", output.String())
}

func TestAnsiWritesIndentedSimple(t *testing.T) {
	input := bytes.Buffer{}
	input.WriteString("Hi there\nAnd indent this too.")

	output := bytes.Buffer{}

	w := NewAnsiScreenWriter(&input, &output)
	w.PrefixEachLine("  ")
	w.Process()

	assert.Equal(t, "  Hi there\n  And indent this too.", output.String())
}

func TestAnsiWritesIndentedWithCR(t *testing.T) {
	input := bytes.Buffer{}
	input.WriteString("Hi there\nAnd indent this too.\rShould still be indented.")

	output := bytes.Buffer{}

	w := NewAnsiScreenWriter(&input, &output)
	w.PrefixEachLine("  ")
	w.Process()

	assert.Equal(t, "  Hi there\n  And indent this too.\r  Should still be indented.", output.String())
}

func TestAnsiWritesIndentedWithCursorUp(t *testing.T) {
	input := bytes.Buffer{}
	input.WriteString("Hi there\nAnd indent this too.\033[1AIndent again.")

	output := bytes.Buffer{}

	w := NewAnsiScreenWriter(&input, &output)
	w.PrefixEachLine("  ")
	w.Process()

	assert.Equal(t, "  Hi there\n  And indent this too.\033[1A  Indent again.", output.String())
}
