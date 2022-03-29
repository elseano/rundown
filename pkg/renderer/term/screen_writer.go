package term

import (
	"fmt"
	"io"
	"strings"

	"github.com/elseano/rundown/pkg/util"
)

type ScreenWriter struct {
	beforeOutputHandler Handler
	afterOutputHandler  Handler
	output              io.Writer
	buffer              Buffer
	doneChan            chan bool
}

type Handler func()

type Buffer interface {
	String() (string, int)
	SubscribeToFlush() chan bool
}

func NewScreenWriter(writer io.Writer, buffer Buffer) *ScreenWriter {
	return &ScreenWriter{
		output:   writer,
		buffer:   buffer,
		doneChan: make(chan bool),
	}
}

func (s *ScreenWriter) SetBeforeFunc(handler Handler) {
	s.beforeOutputHandler = handler
}

func (s *ScreenWriter) SetAfterFunc(handler Handler) {
	s.afterOutputHandler = handler
}

func (s *ScreenWriter) Write(b []byte) (int, error) {
	if s.beforeOutputHandler != nil {
		s.beforeOutputHandler()
	}

	i, err := s.output.Write(b)

	if s.afterOutputHandler != nil {
		s.afterOutputHandler()
	}

	return i, err
}

func (s *ScreenWriter) Run() {
	lastLinesWritten := -1
	lastWrite := ""

	writer := func() {
		output, lines := s.buffer.String()

		util.Logger.Debug().Msgf("Rendering: %#v", output)

		if output != lastWrite {
			if lastLinesWritten > 1 {
				s.output.Write([]byte(fmt.Sprintf("\033[%dA", lastLinesWritten-1)))
			}

			s.output.Write([]byte{'\r'})

			// Convert output newlines to \r\n because we're normally in RAW mode, and Windows.
			output = strings.ReplaceAll(output, "\n", "\r\n")

			// If the last line isn't empty, then we don't want to go back up one line on redraw.
			if !strings.HasSuffix(output, "\r\n") {
				lines--
			}

			// s.Write([]byte(fmt.Sprintf("FLUSH %d\n\r", lastLinesWritten)))

			s.Write([]byte(output))
			// s.Write([]byte(fmt.Sprintf("%#v", output)))
			lastLinesWritten = lines
		}
	}

	flush := s.buffer.SubscribeToFlush()

	for {
		select {
		case <-s.doneChan:
			writer()
			return
		case <-flush:
			writer()
		}
	}
}

func (s *ScreenWriter) Done() {
	s.doneChan <- true
}
