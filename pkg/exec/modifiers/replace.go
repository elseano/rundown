package modifiers

import (
	"bytes"

	"github.com/elseano/rundown/pkg/exec/scripts"
)

type Replace struct {
	ExecutionModifier

	env map[string]string
}

func NewReplace(env map[string]string) *Replace {
	return &Replace{
		ExecutionModifier: &NullModifier{},
		env:               env,
	}
}

func (m *Replace) PrepareScripts(scripts *scripts.ScriptManager) {
	script := scripts.GetBase()

	if script.ShellScript {
		for key, value := range m.env {
			script.Contents = bytes.ReplaceAll(script.Contents, []byte(key), []byte(value))
		}
	}
}
