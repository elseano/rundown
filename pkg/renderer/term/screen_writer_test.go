package term

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestBuffer struct {
	Flushes   []string
	flushChan chan bool
	cFlush    int
}

func (t *TestBuffer) SubscribeToFlush() chan bool {
	t.flushChan = make(chan bool)
	return t.flushChan
}

func (t *TestBuffer) String() (string, int) {
	thisFlush := t.Flushes[t.cFlush]
	lines := strings.Split(thisFlush, "\n")

	fmt.Printf("Returning %#v\n", thisFlush)

	t.cFlush++

	return thisFlush, len(lines)
}

func (t *TestBuffer) Flush() {
	if t.flushChan != nil {
		t.flushChan <- true
	}
}

func TestScreenWriterDuplicateRendering(t *testing.T) {
	output := &bytes.Buffer{}
	input := &TestBuffer{
		Flushes: []string{
			"One",
			"One\nTwo\n",
			"One\nTwo - Three\n",
		},
	}

	w := NewScreenWriter(output, input)
	go w.Run()

	time.Sleep(100 * time.Millisecond)

	input.Flush()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "\rOne", output.String())
	output.Reset()

	input.Flush()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "\rOne\r\nTwo\r\n", output.String())
	output.Reset()

	input.Flush()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "\033[2A\rOne\r\nTwo - Three\r\n", output.String())
}
