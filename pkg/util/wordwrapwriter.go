package util

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

type WordWrapWriter struct {
	io.Writer
	AfterWrap      func(writer io.Writer) int
	CurrentLinePos int
	Columns        int
	lastAfterWrap  int
}

func NewWordWrapWriter(out io.Writer, columns int) *WordWrapWriter {
	return &WordWrapWriter{
		Writer:  out,
		Columns: columns,
	}
}

func (w *WordWrapWriter) Reset() {
	w.CurrentLinePos = 0
	w.lastAfterWrap = 0
}

func (w *WordWrapWriter) SetAfterWrap(wrapper func(writer io.Writer) int) {
	w.AfterWrap = wrapper
}

func (w *WordWrapWriter) trueLength(b []byte) int {
	noc := RemoveColors(string(b))

	return utf8.RuneCountInString(strings.TrimSuffix(noc, "\n"))
}

var wordSplitter = regexp.MustCompile("([\r ])?([^\r ]*)")

func (w *WordWrapWriter) Write(b []byte) (n int, err error) {
	lines := bytes.SplitAfter(b, []byte("\n"))
	for _, line := range lines {
		words := wordSplitter.FindAllSubmatch(line, -1)
		for _, word := range words {
			spaceBytes := word[1]
			wordBytes := word[2]

			trueLength := w.trueLength(wordBytes)

			// We are returning to the start of the line? Reset positions and treat as a wrap.
			// Except if it's CRLF.
			if bytes.Equal(spaceBytes, []byte("\r")) && !bytes.Equal(wordBytes, []byte("\n")) {
				w.Writer.Write([]byte("\r"))
				w.CurrentLinePos = 0
				if w.AfterWrap != nil {
					w.lastAfterWrap = w.AfterWrap(w.Writer)
					w.CurrentLinePos = w.lastAfterWrap
				} else {
					w.lastAfterWrap, w.CurrentLinePos = 0, 0
				}
				spaceBytes = []byte{}
			} else if trueLength+w.CurrentLinePos+len(spaceBytes) > w.Columns {
				// Wrap if the word will spill over
				w.Writer.Write([]byte("\n"))
				w.CurrentLinePos = 0
				if w.AfterWrap != nil {
					w.lastAfterWrap = w.AfterWrap(w.Writer)
					w.CurrentLinePos = w.lastAfterWrap
				} else {
					w.lastAfterWrap, w.CurrentLinePos = 0, 0
				}
				spaceBytes = []byte{}
			}

			wn, err := w.Writer.Write(spaceBytes)
			n = n + wn
			if err != nil {
				return n, err
			}

			wn, err = w.Writer.Write(wordBytes)
			n = n + wn
			if err != nil {
				return n, err
			}

			w.CurrentLinePos = w.CurrentLinePos + trueLength + len(spaceBytes) // +1 for the space.
		}

		// ln, err := w.Writer.Write([]byte("\r\n"))
		// n = n + ln
		// if err != nil {
		// 	return n, err
		// }
		if bytes.HasSuffix(line, []byte("\n")) {
			w.CurrentLinePos = 0
		}
	}

	return len(b), nil
}
