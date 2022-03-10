package testutil

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertLines(t *testing.T, expected string, actual string) bool {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	for i := 0; i < len(expectedLines); i = i + 1 {
		if i > len(actualLines)-1 {
			t.Fatalf("Expected %s, but got no line", expectedLines[i])
		} else if strings.TrimSpace(expectedLines[i]) == "" && strings.TrimSpace(actualLines[i]) == "" {
			continue
		} else if !assert.Equal(t, strings.TrimRight(expectedLines[i], " "), strings.TrimRight(actualLines[i], " "), "Mismatch on line "+strconv.Itoa(i)) {
			b := strings.Builder{}
			for index, line := range strings.Split(actual, "\n") {
				marker := " "
				if index == i {
					marker = "*"
				}
				b.WriteString(fmt.Sprintf("%s %00d: %s\n", marker, index+1, line))
			}
			t.Logf("Actual output:\n%s", b.String())
			break
		}
	}

	if len(expectedLines) != len(actualLines) {
		t.Fatalf("Expected %d lines, but got %d lines", len(expectedLines), len(actualLines))
	}

	return t.Failed()
}

type TestWriter struct {
	t *testing.T
}

func (tw TestWriter) Write(p []byte) (n int, err error) {
	tw.t.Logf(string(p))
	return len(p), nil
}

func NewTestWriter(t *testing.T) TestWriter {
	return TestWriter{t}
}

func NewTestLogger(t *testing.T) *log.Logger {
	return log.New(TestWriter{t}, "TEST: ", log.Ltime)
}
