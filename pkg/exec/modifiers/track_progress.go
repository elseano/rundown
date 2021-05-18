package modifiers

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
)

type TrackProgress struct {
	ExecutionModifier
	signallingKey string
	wait          sync.WaitGroup
	startedAt     time.Time
	endedAt       time.Time
	exitCode      int
}

func NewTrackProgress() *TrackProgress {
	return &TrackProgress{
		ExecutionModifier: &NullModifier{},
	}
}

var progressScript = "echo %s START >> $RDRPC\n$%s\nEC=$?;echo %s END $EC >> $RDRPC;exit $EC"

func (m *TrackProgress) PrepareScripts(scripts *scripts.ScriptManager) {
	m.signallingKey = "ABC123"

	script := scripts.GetPrevious()

	scripts.AddScript("PROGRESS", "/bin/bash", []byte(fmt.Sprintf(progressScript, m.signallingKey, script.EnvReferenceName, m.signallingKey)))

	bus.Subscribe(m)
}

func (m *TrackProgress) GetResult() []ModifierResult {
	bus.Unsubscribe(m)

	if !m.endedAt.IsZero() {
		return []ModifierResult{
			{
				Key:   "Duration",
				Value: m.endedAt.Sub(m.startedAt),
			},
			{
				Key:   "ExitCode",
				Value: m.exitCode,
			},
		}
	}

	return nil
}

func (m *TrackProgress) ReceiveEvent(event bus.Event) {
	if rpcEvent, ok := event.(*rpc.RpcMessage); ok {
		data := rpcEvent.Data

		if data == m.signallingKey+" START" {
			m.wait.Add(1)
			m.startedAt = time.Now()

		} else if strings.HasPrefix(data, m.signallingKey+" END") {

			m.wait.Done()
			m.endedAt = time.Now()
			exitCodeString := data[len(m.signallingKey+" END "):]
			if code, err := strconv.Atoi(exitCodeString); err == nil {
				m.exitCode = code
			}
		}
	}
}
