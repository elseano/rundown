package exec

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/elseano/rundown/pkg/exec/modifiers"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/util"
)

/**
* Modifiers:
* 	nospin - No Spinner
*		interactive - Support STDIN
*		skip_on_success - Skip to next heading when code executes successfully
*		skip_on_failure - Skip to next...
*   stdout - Display stdout
*   stderr - Display stderr
*   reveal - Render the code block before running it
*   norun - Show the code, don't run it
*   named - The first line is a comment marker, space, then text for the spinner
*   abort - Implies named and stdout. Displays a named execution failure, and exits without showing the script.
 */

type ExecutionIntent struct {
	Via            string
	Script         []byte
	ReplaceProcess bool

	subscriptions  []rpc.Subscription
	terminationKey string
	modifiers      *modifiers.ExecutionModifiers
	env            map[string]string
	cwd            string
}

type ExecutionResult struct {
	Output   []byte
	ExitCode int
	Env      map[string]string
}

func NewExecution(via string, script []byte, cwd string) (*ExecutionIntent, error) {
	intent := &ExecutionIntent{
		Via:            via,
		Script:         script,
		subscriptions:  []rpc.Subscription{},
		terminationKey: "ABC123",
		modifiers:      modifiers.NewExecutionModifiers(),
		env:            map[string]string{},
		cwd:            cwd,
	}

	return intent, nil
}

func (i *ExecutionIntent) ImportEnv(env map[string]string) {
	for k, v := range env {
		i.env[k] = v
	}
}

func (i *ExecutionIntent) AddModifier(mod modifiers.ExecutionModifier) {
	i.modifiers.AddModifier(mod)
}

func (i *ExecutionIntent) Execute() (*ExecutionResult, error) {
	var baseEnv = map[string]string{}
	for k, v := range i.env {
		baseEnv[k] = v
	}

	// If we're replacing the process, we don't need to worry about the RPC interface.
	if !i.ReplaceProcess {
		rpcEndpoint, err := rpc.Start()
		if err != nil {
			return nil, err
		}

		defer rpcEndpoint.Close()

		baseEnv[rpc.EnvironmentVariableName] = rpcEndpoint.Path
	}

	var content = scripts.NewScriptManager()
	content.SetBaseScript(i.Via, i.Script)
	defer content.RemoveAll()

	// var modRunners = modifiers.NewExecutionModifiers()
	// i.modifiers.AddModifier(modifiers.NewEnvironmentCapture())
	// modRunners.AddModifier(modifiers.NewTrackProgress())
	// modRunners.AddModifier(modifiers.NewStdout())

	util.Logger.Trace().Msgf("Process: %s", i.Via)
	util.Logger.Trace().Msgf("Script: %s", i.Script)

	i.modifiers.PrepareScripts(content)

	// modRunners.PrepareScripts(content)

	if i.ReplaceProcess {
		lastScript, err := content.Write()

		if err != nil {
			return nil, err
		}

		simpleEnv := make([]string, len(baseEnv))
		i := 0
		for k, v := range baseEnv {
			simpleEnv[i] = fmt.Sprintf("%s=%s", k, v)
		}

		if err = syscall.Exec(lastScript.AbsolutePath, []string{lastScript.AbsolutePath}, simpleEnv); err != nil {
			return nil, err
		}
	}

	err, process, stdout := launchProcess(content, baseEnv, i.cwd)

	if err != nil {
		return nil, err
	}

	var outputCaptureGroup = []io.Writer{}
	var waiter sync.WaitGroup

	outputCaptureGroup = i.modifiers.GetStdout()
	captureOutputStream(outputCaptureGroup, &waiter, stdout)

	util.Logger.Trace().Msg("Waiting process termination")
	waitErr := process.Wait()
	exitCode := determineExitCode(waitErr)

	util.Logger.Trace().Msg("Waiting goroutine termination")
	waiter.Wait()

	execResult := &ExecutionResult{
		ExitCode: exitCode,
	}

	results := i.modifiers.GetResult(exitCode)

	for _, result := range results {
		switch result.Key {
		case "Env":
			execResult.Env = result.Value.(map[string]string)
		case "Duration":
			util.Logger.Trace().Dur("Time", result.Value.(time.Duration)).Msg("Timing data available")
		case "Output":
			execResult.Output = result.Value.([]byte)
		}
	}

	util.Logger.Trace().Msgf("Results: %#v", execResult)

	return execResult, nil
}

func launchProcess(content *scripts.ScriptManager, baseEnv map[string]string, cwd string) (error, *Process, *io.PipeReader) {
	lastScript, err := content.Write()

	if err != nil {
		return err, nil, nil
	}

	cmd := exec.Command(lastScript.AbsolutePath)
	cmd.Dir = cwd

	for name, val := range baseEnv {
		cmd.Env = append(cmd.Env, name+"="+val)
	}

	for name, val := range content.GenerateReferences() {
		cmd.Env = append(cmd.Env, name+"="+val)
	}

	util.Logger.Trace().Msg("Launching process...")

	process := NewProcess(cmd)
	stdout, err := process.Start()

	if err == nil {
		util.Logger.Trace().Msg("Process started ok")
	}

	return err, process, stdout
}

func captureOutputStream(outputCaptureGroup []io.Writer, waiter *sync.WaitGroup, stdout *io.PipeReader) {
	util.Logger.Trace().Msgf("Copying STDOUT to %d other streams: %#v", len(outputCaptureGroup), outputCaptureGroup)

	var writer = io.MultiWriter(outputCaptureGroup...)

	waiter.Add(1)
	go func() {
		util.Logger.Trace().Msg("Capturing STDOUT")
		w, _ := io.Copy(writer, stdout)
		util.Logger.Trace().Msgf("STDOUT closed, bytes written: %d", w)

		waiter.Done()
	}()
}

func determineExitCode(waitErr error) int {
	exitCode := 0

	if exitErr, ok := waitErr.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	util.Logger.Trace().Msgf("Terminated with %d", exitCode)

	return exitCode
}
