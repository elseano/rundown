package modifiers

import (
	"io"
	"os"

	"github.com/elseano/rundown/pkg/bus"
)

type StdoutStream struct {
	bus.Handler
	ExecutionModifier
	writer *os.File
	Reader *os.File
}

func NewStdoutStream() *StdoutStream {
	reader, writer, _ := os.Pipe()

	return &StdoutStream{
		ExecutionModifier: &NullModifier{},
		Reader:            reader,
		writer:            writer,
	}
}

func (m *StdoutStream) GetStdout() []io.Writer {
	return []io.Writer{m.writer}
}

func (m *StdoutStream) GetResult(int) []ModifierResult {
	m.writer.Close()
	return []ModifierResult{}
}
