package exec

import (
	"fmt"
	"io"
	"strings"
	"time"

	go_exec "os/exec"

	"github.com/elseano/rundown/pkg/exec/scripts"
	rdutil "github.com/elseano/rundown/pkg/util"
)

type Runner struct {
	Script *scripts.Script
	env    map[string]string
}

func NewRunner() *Runner {
	return &Runner{env: map[string]string{}}
}

func (r *Runner) SetScript(binaryPath string, language string, source []byte) (*scripts.Script, error) {
	script, err := scripts.NewScript(binaryPath, language, source)

	rdutil.Logger.Debug().Msgf("Script created. Binary: %s, Command Line: %s", script.BinaryPath, script.CommandLine)

	if err != nil {
		return nil, err
	}

	r.Script = script

	return r.Script, nil
}

func (r *Runner) ImportEnv(env map[string]string) {
	for k, v := range env {
		r.env[k] = v
	}
}

// Executes the script, replacing the current process.
func (r *Runner) ExecuteAndReplace() error {

	// 	wrapperScript := scripts.NewScript()
	// go_exec.Command(wrapperScript.)

	return nil
}

type Running struct {
	Runner       *Runner
	Stdout       io.ReadCloser
	cmd          *go_exec.Cmd
	startedAt    time.Time
	StderrOutput []byte
	Stderr       io.ReadCloser
}

func (r *Runner) Prepare() (*Running, error) {
	wrapperScriptContents := r.Script.CommandLine

	// Replace $SCRIPT_FILE otherwise it appears on the command line twice.
	wrapperScriptContents = strings.ReplaceAll(wrapperScriptContents, "$SCRIPT_FILE", r.Script.AbsolutePath)

	wrapperScript, err := scripts.NewScript("bash", "bash", []byte(wrapperScriptContents))
	if err != nil {
		return nil, err
	}
	wrapperScript.Write()
	wrapperScript.MakeExecutable()

	rdutil.Logger.Debug().Msgf("Wrapper script is %s", wrapperScript.AbsolutePath)
	rdutil.Logger.Debug().Msgf("Provided script is %s", r.Script.AbsolutePath)

	cmd := go_exec.Command(wrapperScript.BinaryPath, wrapperScript.AbsolutePath)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	return &Running{
		Runner: r,
		cmd:    cmd,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

func (r *Running) Start() error {
	if err := r.Runner.Script.Write(); err != nil {
		return err
	}

	r.cmd.Args = append(r.cmd.Args, r.Runner.Script.AbsolutePath)
	for k, v := range r.Runner.env {
		r.cmd.Env = append(r.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	r.cmd.Env = append(r.cmd.Env, fmt.Sprintf("SCRIPT_FILE=%s", r.Runner.Script.AbsolutePath))
	r.cmd.Dir = r.Runner.env["PWD"]

	rdutil.Logger.Debug().Msgf("Running with environment: %+v", r.cmd.Env)
	rdutil.Logger.Debug().Msgf("Running in path: %s", r.cmd.Dir)

	r.startedAt = time.Now()
	err := r.cmd.Start()
	if err != nil && r.cmd.Process == nil {
		rdutil.Logger.Debug().Msgf("Error starting script: %#v", err)
		return err
	}

	return nil
}

func (r *Running) Wait() (int, time.Duration, error) {
	err := r.cmd.Wait()

	if err != nil {
		if exitErr, ok := err.(*go_exec.ExitError); ok {
			rdutil.Logger.Debug().Msgf("Process exited with %d", r.cmd.ProcessState.ExitCode())
			r.StderrOutput = exitErr.Stderr
			return exitErr.ExitCode(), time.Since(r.startedAt), nil
		}

		return 0, 0, err
	}

	rdutil.Logger.Debug().Msgf("Process exited with %d", r.cmd.ProcessState.ExitCode())

	return r.cmd.ProcessState.ExitCode(), time.Since(r.startedAt), nil
}
