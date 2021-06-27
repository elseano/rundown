package modifiers

import (
	"regexp"
	"strings"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/spinner"
)

// SpinnerFromScript modifier requires the TrackProgress modifier.
type SpinnerFromScript struct {
	bus.Handler
	ExecutionModifier
	Spinner                 spinner.Spinner
	CommentsAsSpinnerTitles bool
}

func NewSpinnerFromScript() *SpinnerFromScript {
	return &SpinnerFromScript{
		ExecutionModifier: &NullModifier{},
	}
}

var commentDetector = regexp.MustCompile(`#+\s+(.*)`)
var spinnerRpcUpdate = []byte("echo SETSPINNER $1 >> $$RDRPC")

func (m *SpinnerFromScript) PrepareScripts(scripts *scripts.ScriptManager) {
	if m.CommentsAsSpinnerTitles {
		script := scripts.GetBase()

		if script.ShellScript {
			script.Contents = commentDetector.ReplaceAll(script.Contents, spinnerRpcUpdate)
		}
	}

	bus.Subscribe(m)
}

func (m *SpinnerFromScript) GetResult(int) []ModifierResult {
	bus.Unsubscribe(m)

	return []ModifierResult{}
}

func (m *SpinnerFromScript) ReceiveEvent(event bus.Event) {
	if rpcEvent, ok := event.(*rpc.RpcMessage); ok {
		data := rpcEvent.Data

		if strings.HasPrefix(data, "SETSPINNER ") {
			spinnerTitle := data[len("SETSPINNER "):]

			bus.Emit(&spinner.ChangeTitleEvent{Title: spinnerTitle})
		}
	}
}
