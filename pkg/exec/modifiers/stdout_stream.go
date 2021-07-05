package modifiers

import (
	"io"

	"github.com/elseano/rundown/pkg/bus"
)

type StdoutStream struct {
	bus.Handler
	ExecutionModifier
	destination io.Writer
}

func NewStdoutStream(destination io.Writer) *StdoutStream {
	return &StdoutStream{
		ExecutionModifier: &NullModifier{},
		destination:       destination,
	}
}

func (m *StdoutStream) GetStdout() []io.Writer {
	return []io.Writer{m.destination}
}
