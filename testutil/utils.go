package testutil

import (
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertLines(t *testing.T, expected string, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	for i := 0; i < len(expectedLines); i = i + 1 {
		if i > len(actualLines)-1 {
			t.Fatalf("Expected %s, but got no line", expectedLines[i])
		} else if strings.TrimSpace(expectedLines[i]) == "" && strings.TrimSpace(actualLines[i]) == "" {
			continue
		} else if !assert.Equal(t, expectedLines[i], actualLines[i], "Mismatch on line "+strconv.Itoa(i)) {
			break
		}
	}

	if len(expectedLines) != len(actualLines) {
		t.Fatalf("Expected %d lines, but got %d lines", len(expectedLines), len(actualLines))
	}
}

type TestWriter struct {
	t *testing.T
}

func (tw TestWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

func NewTestLogger(t *testing.T) *log.Logger {
	return log.New(TestWriter{t}, "TEST: ", log.Ltime)
}
