package modifiers

import (
	"strings"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/util"
)

type EnvironmentCapture struct {
	ExecutionModifier

	capturedEnv   map[string]string
	capturing     bool
	capturedTimes int
}

func NewEnvironmentCapture() *EnvironmentCapture {
	return &EnvironmentCapture{
		ExecutionModifier: &NullModifier{},
	}
}

var envDumpCommand = []byte("\n\necho ENVDUMP >> $RDRPC;env >> $RDRPC;echo '' >> $RDRPC\n\n")

func (m *EnvironmentCapture) PrepareScripts(scripts *scripts.ScriptManager) {
	script := scripts.GetBase()

	if script.ShellScript {
		script.Contents = append(envDumpCommand, script.Contents...)
		script.Contents = append(script.Contents, envDumpCommand...)

		bus.Subscribe(m)
	}
}

func (m *EnvironmentCapture) GetResult(int) []ModifierResult {
	bus.Unsubscribe(m)

	if m.capturedTimes > 1 {
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

			util.Logger.Trace().Str("key", set[0]).Str("val", set[1]).Msg("Consuming variable")

			if len(set) > 1 {
				if val, ok := m.capturedEnv[set[0]]; ok {
					util.Logger.Trace().Msg("Key already present, checking...")

					if val == set[1] {
						util.Logger.Trace().Msg("Value of key is the same. Deleting.")

						delete(m.capturedEnv, set[0])
					} else {
						util.Logger.Trace().Msg("Value of key is different. Storing.")
						m.capturedEnv[set[0]] = set[1]
					}
				} else {
					util.Logger.Trace().Msg("Key is new. Storing.")
					m.capturedEnv[set[0]] = set[1]
				}
			}

		}
	}
}
