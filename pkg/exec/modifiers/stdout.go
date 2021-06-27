package modifiers

import (
	"bytes"
	"io"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/scripts"
)

type Stdout struct {
	bus.Handler
	ExecutionModifier
	buffer        bytes.Buffer
	scriptManager *scripts.ScriptManager
}

func NewStdout() *Stdout {
	return &Stdout{
		ExecutionModifier: &NullModifier{},
	}
}

func (m *Stdout) PrepareScripts(scripts *scripts.ScriptManager) {
	m.scriptManager = scripts
}

func (m *Stdout) GetResult(int) []ModifierResult {
	output := m.buffer.Bytes()
	output = bytes.ReplaceAll(output, []byte("\r\n"), []byte("\n"))                                     // Terminal is in RAW mode, convert back to unix line feeds.
	output = bytes.ReplaceAll(output, []byte(m.scriptManager.GetBase().AbsolutePath), []byte("SCRIPT")) // Mask the temp file name with SCRIPT

	return []ModifierResult{
		{
			Key:   "Output",
			Value: output,
		},
	}
}

func (m *Stdout) GetStdout() []io.Writer {
	return []io.Writer{&m.buffer}
}
