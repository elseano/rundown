package modifiers

import (
	"fmt"
	"strings"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/util"
)

type EnvironmentCapture struct {
	ExecutionModifier

	Variables []string

	capturedEnv   map[string]string
	capturing     bool
	capturedTimes int
}

func NewEnvironmentCapture(variables []string) *EnvironmentCapture {
	return &EnvironmentCapture{
		ExecutionModifier: &NullModifier{},
		Variables:         variables,
	}
}

var envDumpCommand = "\n\necho ENVDUMP >> $RDRPC;CAPTURE;echo '' >> $RDRPC\n\n"

func (m *EnvironmentCapture) PrepareScripts(scripts *scripts.ScriptManager) {
	script := scripts.GetBase()

	if script.ShellScript {
		envs := make([]string, len(m.Variables))
		for i, variable := range m.Variables {
			envs[i] = fmt.Sprintf("echo \"%s=${%s}\" >> $RDRPC", variable, variable)
		}

		envCaptureCommand := strings.Replace(envDumpCommand, "CAPTURE", strings.Join(envs, ";"), 1)

		script.Suffix = []byte(envCaptureCommand)

		bus.Subscribe(m)
	}
}

func (m *EnvironmentCapture) GetResult(int) []ModifierResult {
	bus.Unsubscribe(m)

	if m.capturedEnv != nil {
		return []ModifierResult{
			{
				Key:   "Env",
				Value: m.capturedEnv,
			},
		}
	}

	return nil
}

func (m *EnvironmentCapture) ReceiveEvent(event bus.Event) {
	if rpcEvent, ok := event.(*rpc.RpcMessage); ok {

		data := rpcEvent.Data

		switch data {
		case "ENVDUMP":
			util.Logger.Trace().Msg("Entering environment capture mode.")

			m.capturing = true

			if m.capturedEnv == nil {
				m.capturedEnv = map[string]string{}
			}

			m.capturedTimes += 1

		case "":
			if m.capturing {
				util.Logger.Trace().Msg("Exiting environment capture mode.")

				m.capturing = false
			}

		default:

			set := strings.SplitN(data, "=", 2)

			if len(set) < 2 {
				return // Message probably not for us.
			}

			util.Logger.Trace().Msgf("Received var %s: %s.", set[0], set[1])

			if len(set) > 1 {
				util.Logger.Trace().Str("key", set[0]).Str("val", set[1]).Msg("Consuming variable")
				m.capturedEnv[set[0]] = set[1]
			}

		}
	}
}
