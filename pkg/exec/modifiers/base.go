package modifiers

import (
	"io"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/scripts"
)

type ModifierResult struct {
	Key   string
	Value interface{}
}

type ExecutionModifier interface {
	PrepareScripts(scripts *scripts.ScriptManager)
	GetResult() []ModifierResult
	GetStdout() []io.Writer
}

type ExecutionModifierList []ExecutionModifier

type ExecutionModifiers struct {
	ExecutionModifier
	mods []ExecutionModifier
}

func NewExecutionModifiers() *ExecutionModifiers {
	return &ExecutionModifiers{
		mods: []ExecutionModifier{},
	}
}

func (m *ExecutionModifiers) AddModifier(modifier ExecutionModifier) {
	m.mods = append(m.mods, modifier)
}

func (m *ExecutionModifiers) PrepareScripts(scripts *scripts.ScriptManager) {
	for _, m := range m.mods {
		m.PrepareScripts(scripts)
	}
}

func (m *ExecutionModifiers) GetStdout() []io.Writer {
	output := []io.Writer{}

	for _, m := range m.mods {
		output = append(output, m.GetStdout()...)
	}

	return output
}

func (m *ExecutionModifiers) GetResult() []ModifierResult {
	var results = []ModifierResult{}

	for _, m := range m.mods {
		if result := m.GetResult(); len(result) > 0 {
			results = append(results, result...)
		}
	}

	return results
}

type NullModifier struct {
	ExecutionModifier
}

func (m *NullModifier) GetStdout() []io.Writer {
	return []io.Writer{}
}

func (m *NullModifier) PrepareScripts(scripts *scripts.ScriptManager) {
}

func (m *NullModifier) ReceiveEvent(event bus.Event) {
}

func (m *NullModifier) GetResult() []ModifierResult {
	return []ModifierResult{}
}
