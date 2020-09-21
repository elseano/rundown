package util

import (
	"io"
	"bytes"
	"regexp"
)

type WordWrapWriter struct {
	io.Writer
	AfterWrap func(writer io.Writer) int
	CurrentLinePos int
	Columns int
	lastAfterWrap int
}

func NewWordWrapWriter(out io.Writer, columns int) *WordWrapWriter {
	return &WordWrapWriter{
		Writer: out,
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
	return len(RemoveColors(string(b)))
}

var wordSplitter = regexp.MustCompile("([\r ])?([^\r ]*)")

func (w *WordWrapWriter) Write(b []byte) (n int, err error) {
	for _, line := range bytes.SplitAfter(b, []byte("\n")) {
		for _, word := range wordSplitter.FindAllSubmatch(line, -1) {
			spaceBytes := word[1]
			wordBytes := word[2]

			trueLength := w.trueLength(wordBytes)

			// We are returning to the start of the line? Reset positions and treat as a wrap.
			if bytes.Equal(spaceBytes, []byte("\r")) {
				w.Writer.Write([]byte("\r"))
				w.CurrentLinePos = 0
				if w.AfterWrap != nil {
					w.lastAfterWrap = w.AfterWrap(w.Writer)
					w.CurrentLinePos = w.lastAfterWrap
				} else {
					w.lastAfterWrap, w.CurrentLinePos = 0,0
				}
				spaceBytes = []byte{}
			} else if trueLength + w.CurrentLinePos + len(spaceBytes) > w.Columns {
				// Wrap if the word will spill over
				w.Writer.Write([]byte("\n"))
				w.CurrentLinePos = 0
				if w.AfterWrap != nil {
					w.lastAfterWrap = w.AfterWrap(w.Writer)
					w.CurrentLinePos = w.lastAfterWrap
				} else {
					w.lastAfterWrap, w.CurrentLinePos = 0,0
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

	return n, nil
}

