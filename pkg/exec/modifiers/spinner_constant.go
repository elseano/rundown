package modifiers

import (
	"os"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/spinner"
)

// SpinnerFromScript modifier requires the TrackProgress modifier.
type SpinnerConstant struct {
	bus.Handler

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

	bus.Subscribe(m)
}

func (m *SpinnerConstant) GetResult(exitCode int) []ModifierResult {
	bus.Unsubscribe(m)

	// if exitCode == 0 {
	// 	m.Spinner.Success("OK")
	// } else {
	// 	m.Spinner.Error(fmt.Sprintf("Exit code %d", exitCode))
	// }

	return []ModifierResult{}
}

func (m *SpinnerConstant) ReceiveEvent(event bus.Event) {
	if rpcEvent, ok := event.(*rpc.RpcMessage); ok {
		data := rpcEvent.Data

		if data == "STOPSPINNER" {
			m.Spinner.Stop()
		} else if data == "STARTSPINNER" {
			m.Spinner.Start()
		}
	}
}
