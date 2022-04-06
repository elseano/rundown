package modifiers

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/util"
)

// SpinnerFromScript modifier requires the TrackProgress modifier.
type SpinnerFromScript struct {
	bus.Handler
	ExecutionModifier
	Spinner                 *SpinnerConstant
	CommentsAsSpinnerTitles bool
	nameHasBeenSet          bool
}

func NewSpinnerFromScript(CommentsAsSpinnerTitles bool, spinner *SpinnerConstant) *SpinnerFromScript {
	return &SpinnerFromScript{
		ExecutionModifier:       &NullModifier{},
		Spinner:                 spinner,
		CommentsAsSpinnerTitles: CommentsAsSpinnerTitles,
	}
}

var commentDetector = regexp.MustCompile(`#+\s+(.*)`)

const marker = "R;SETSPINNER "

func (m *SpinnerFromScript) PrepareScripts(scripts *scripts.ScriptManager) {
	if m.CommentsAsSpinnerTitles {
		script := scripts.GetBase()

		if script.ShellScript {
			script.Contents = commentDetector.ReplaceAllFunc(script.Contents, func(b []byte) []byte {
				matches := commentDetector.FindAllSubmatch(b, 1)

				// Base64 encode
				result := base64.StdEncoding.EncodeToString(matches[0][1])
				return []byte(fmt.Sprintf("echo -n -e \"\x1b]R;SETSPINNER %s\x9c\"", result))
			})
		}
	}

	bus.Subscribe(m)
}

func (m *SpinnerFromScript) GetResult(int) []ModifierResult {
	bus.Unsubscribe(m)

	return []ModifierResult{}
}

type OSCHandler interface {
	HandleOSC(string) bool
}

func (m *SpinnerFromScript) HandleOSC(osc string) bool {
	if strings.HasPrefix(osc, marker) {
		title64 := osc[len(marker):]
		title, _ := base64.StdEncoding.DecodeString(title64)
		spinnerTitle := string(title)

		if m.nameHasBeenSet {
			m.Spinner.Spinner.NewStep(spinnerTitle)
		} else {
			m.Spinner.Spinner.SetMessage(spinnerTitle)
			m.nameHasBeenSet = true
		}

		return true
	}

	return false
}

func (m *SpinnerFromScript) ReceiveEvent(event bus.Event) {
	if rpcEvent, ok := event.(*rpc.RpcMessage); ok {
		data := rpcEvent.Data

		if strings.HasPrefix(data, "SETSPINNER ") {
			util.Logger.Debug().Msgf("Setting spinner message")
			spinnerTitle := data[len("SETSPINNER "):]

			if m.nameHasBeenSet {
				m.Spinner.Spinner.NewStep(spinnerTitle)
			} else {
				m.Spinner.Spinner.SetMessage(spinnerTitle)
				m.nameHasBeenSet = true
			}

			// bus.Emit(&spinner.ChangeTitleEvent{Title: spinnerTitle})
		}
	}
}
