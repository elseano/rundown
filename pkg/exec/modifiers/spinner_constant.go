package modifiers

import (
	"fmt"
	"os"

	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/spinner"
)

// SpinnerFromScript modifier requires the TrackProgress modifier.
type SpinnerConstant struct {
	ExecutionModifier
	Spinner     spinner.Spinner
	SpinnerName string
}

func NewSpinnerConstant(name string) *SpinnerConstant {
	return &SpinnerConstant{
		ExecutionModifier: &NullModifier{},
		SpinnerName:       name,
	}
}

func (m *SpinnerConstant) PrepareScripts(scripts *scripts.ScriptManager) {
	m.Spinner = spinner.NewSpinner(0, m.SpinnerName, os.Stdout)
	m.Spinner.Start()
}

func (m *SpinnerConstant) GetResult(exitCode int) []ModifierResult {
	if exitCode == 0 {
		m.Spinner.Success("OK")
	} else {
		m.Spinner.Error(fmt.Sprintf("Exit code %d", exitCode))
	}

	return []ModifierResult{}
}
