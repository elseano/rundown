package modifiers

import (
	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	rdutil "github.com/elseano/rundown/pkg/util"
)

type SpinnerControl interface {
	Active() bool
	Start()
	Stop()
	Success(message string)
	Error(message string)
	Skip()
	SetMessage(message string)
	NewStep(message string)
	HideAndExecute(f func())
	CurrentHeading() string
	StampShadow()
}

// SpinnerFromScript modifier requires the TrackProgress modifier.
type SpinnerConstant struct {
	bus.Handler

	ExecutionModifier
	Spinner     SpinnerControl
	SpinnerName string
}

func NewSpinnerConstant(name string, spinner SpinnerControl) *SpinnerConstant {
	return &SpinnerConstant{
		ExecutionModifier: &NullModifier{},
		SpinnerName:       name,
		Spinner:           spinner,
	}
}

func (m *SpinnerConstant) PrepareScripts(scripts *scripts.ScriptManager) {
	rdutil.Logger.Debug().Msg("Spinner start")

	m.Spinner.SetMessage(m.SpinnerName)
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
