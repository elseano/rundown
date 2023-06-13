package exec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
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
	StartedChan    chan interface{}

	subscriptions  []rpc.Subscription
	terminationKey string
	modifiers      *modifiers.ExecutionModifiers
	env            map[string]string
	cwd            string
}

type ExecutionResult struct {
	Scripts  *scripts.ScriptManager
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
		StartedChan:    make(chan interface{}, 1),
	}

	return intent, nil
}

func (i *ExecutionIntent) ImportEnv(env map[string]string) {
	for k, v := range env {
		i.env[k] = v
	}
}

func (i *ExecutionIntent) AddModifier(mod modifiers.ExecutionModifier) modifiers.ExecutionModifier {
	util.Logger.Debug().Msgf("Adding mod: %T", mod)

	i.modifiers.AddModifier(mod)
	return mod
}

func (i *ExecutionIntent) Execute() (*ExecutionResult, error) {
	var baseEnv = map[string]string{}
	for k, v := range i.env {
		baseEnv[k] = v
	}

	var content = scripts.NewScriptManager()
	content.SetBaseScript(i.Via, i.Script)

	defer content.RemoveAll()

	util.Logger.Trace().Msgf("Process: %s", i.Via)
	util.Logger.Trace().Msgf("Script: %s", i.Script)

	i.modifiers.PrepareScripts(content)

	if i.ReplaceProcess {
		lastScript, err := content.Write()

		if err != nil {
			return nil, err
		}

		simpleEnv := os.Environ()

		for k, v := range baseEnv {
			simpleEnv = append(simpleEnv, fmt.Sprintf("%s=%s", k, v))
		}

		for name, val := range content.GenerateReferences() {
			simpleEnv = append(simpleEnv, name+"="+val)
		}

		util.Logger.Trace().Msgf("Replacing process with %s...", lastScript.AbsolutePath)

		if err = syscall.Exec(lastScript.AbsolutePath, []string{lastScript.AbsolutePath}, simpleEnv); err != nil {
			return nil, err
		}
	}

	i.StartedChan <- true

	err, process, stdout, stderr := launchProcess(content, baseEnv, i.cwd)

	if err != nil {
		return nil, err
	}

	var outputCaptureGroup = []io.Writer{}
	var waiter sync.WaitGroup

	outputCaptureGroup = i.modifiers.GetStdout()
	captureOutputStream(outputCaptureGroup, &waiter, stdout, "STDOUT")
	captureOutputStream(outputCaptureGroup, &waiter, stderr, "STDERR")

	util.Logger.Trace().Msg("Waiting process termination")
	waitErr := process.Wait()
	exitCode := determineExitCode(waitErr)

	util.Logger.Trace().Msg("Waiting goroutine termination")
	waiter.Wait()

	execResult := &ExecutionResult{
		ExitCode: exitCode,
		Scripts:  content,
	}

	results := i.modifiers.GetResult(exitCode)

	for _, result := range results {
		switch result.Key {
		case "Env":
			execResult.Env = result.Value.(map[string]string)
		case "Duration":
			util.Logger.Trace().Dur("Time", result.Value.(time.Duration)).Msg("Timing data available")
		case "Output":
			// Trim the filename out of the output.
			output := result.Value.([]byte)

			for _, script := range content.AllScripts() {
				output = bytes.ReplaceAll(output, []byte(fmt.Sprintf("%s: ", script.AbsolutePath)), []byte(""))
				output = bytes.ReplaceAll(output, []byte(script.AbsolutePath), []byte(""))
			}
			execResult.Output = output
		}
	}

	util.Logger.Trace().Msgf("Results: %#v", execResult)

	return execResult, nil
}

func launchProcess(content *scripts.ScriptManager, baseEnv map[string]string, cwd string) (error, *Process, *io.PipeReader, *io.PipeReader) {
	lastScript, err := content.Write()

	if err != nil {
		return err, nil, nil, nil
	}

	cmd := exec.Command(lastScript.AbsolutePath)
	cmd.Dir = cwd

	for name, val := range baseEnv {
		cmd.Env = append(cmd.Env, name+"="+val)
	}

	for name, val := range content.GenerateReferences() {
		util.Logger.Trace().Msgf("ENV %s=%s", name, val)
		cmd.Env = append(cmd.Env, name+"="+val)
	}

	util.Logger.Trace().Fields(map[string]interface{}{"ENV": cmd.Env}).Msg("Launching process...")

	process := NewProcess(cmd)
	stdout, stderr, err := process.Start()

	if err == nil {
		util.Logger.Trace().Msg("Process started ok")
	}

	return err, process, stdout, stderr
}

func captureOutputStream(outputCaptureGroup []io.Writer, waiter *sync.WaitGroup, stdout *io.PipeReader, name string) {
	util.Logger.Trace().Msgf("Copying %s to %d other streams: %#v", name, len(outputCaptureGroup), outputCaptureGroup)

	var writer = io.MultiWriter(outputCaptureGroup...)

	waiter.Add(1)
	go func() {
		defer waiter.Done()

		util.Logger.Trace().Msgf("Capturing %s", name)
		w, err := io.Copy(writer, stdout)
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			util.Logger.Err(err).Msgf("Error: %s", err.Error())
			return
		}
		util.Logger.Trace().Msgf("%s closed, bytes written: %d", name, w)
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
