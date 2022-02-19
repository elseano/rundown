package renderer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/exec/modifiers"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/text"
	"github.com/elseano/rundown/pkg/spinner"
	rutil "github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type RundownConsoleRenderer struct {
	Context *Context
}

func NewRundownConsoleRenderer(context *Context) *RundownConsoleRenderer {
	return &RundownConsoleRenderer{Context: context}
}

func (r *RundownConsoleRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindExecutionBlock, r.renderExecutionBlock)
	reg.Register(ast.KindDescriptionBlock, r.renderNoop)
	reg.Register(ast.KindEnvironmentSubstitution, r.renderEnvironmentSubstitution)
	reg.Register(ast.KindIgnoreBlock, r.renderTodo)
	reg.Register(ast.KindOnFailure, r.renderTodo)
	reg.Register(ast.KindRundownBlock, r.renderTodo)
	reg.Register(ast.KindSaveCodeBlock, r.renderTodo)
	reg.Register(ast.KindSectionEnd, r.renderNoop)
	reg.Register(ast.KindSectionOption, r.renderNoop)
	reg.Register(ast.KindSectionPointer, r.renderNoop)
	reg.Register(ast.KindStopFail, r.renderTodo)
	reg.Register(ast.KindStopOk, r.renderTodo)

	reg.Register(goldast.KindString, r.renderString)
}

func (r *RundownConsoleRenderer) renderNoop(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	return goldast.WalkContinue, nil
}

func (r *RundownConsoleRenderer) renderEnvironmentSubstitution(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if sub, ok := node.(*ast.EnvironmentSubstitution); ok {
		if entering {
			variable := string(sub.Value)
			variable = strings.TrimPrefix(variable, "$")

			if varData, ok := r.Context.Env[variable]; ok {
				w.Write([]byte(varData))
			} else {
				w.Write([]byte(variable))
			}
		}
	}

	return goldast.WalkContinue, nil
}

func (r *RundownConsoleRenderer) renderString(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if str, ok := node.(*goldast.String); ok {
		if entering {
			w.Write(str.Value)
		}
	}

	return goldast.WalkContinue, nil
}

func (r *RundownConsoleRenderer) renderTodo(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		w.WriteString(fmt.Sprintf("TODO - %s", node.Kind().String()))
	}

	return goldast.WalkContinue, nil
}

func (r *RundownConsoleRenderer) renderExecutionBlock(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		return goldast.WalkContinue, nil
	}

	executionBlock := node.(*ast.ExecutionBlock)

	if !executionBlock.Execute {
		return goldast.WalkContinue, nil
	}

	contentReader := text.NewNodeReaderFromSource(executionBlock.CodeBlock, source)

	script, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return goldast.WalkStop, err
	}

	intent, err := exec.NewExecution(executionBlock.With, script, path.Dir(r.Context.RundownFile))
	if err != nil {
		return goldast.WalkStop, err
	}

	intent.ImportEnv(r.Context.Env)

	intent.ReplaceProcess = executionBlock.ReplaceProcess

	// If we're replacing the rundown process, then we don't need to setup spinners, etc.
	// Just execute and check for an error, terminating early.
	if intent.ReplaceProcess {
		_, err := intent.Execute()

		if err != nil {
			return goldast.WalkStop, nil
		}
	}

	// intent.AddModifier(modifiers.NewStdout())
	reader, writer, _ := os.Pipe()

	rutil.Logger.Debug().Msgf("Spinner mode %d", executionBlock.SpinnerMode)

	var spinnerControl spinner.Spinner

	switch executionBlock.SpinnerMode {
	case ast.SpinnerModeInlineAll:
		spinner := modifiers.NewSpinnerConstant(executionBlock.SpinnerName)
		intent.AddModifier(spinner)

		rutil.Logger.Debug().Msg("Inline all mode")
		spinnerDetector := modifiers.NewSpinnerFromScript(true, spinner)
		intent.AddModifier(spinnerDetector)

		spinnerControl = spinner.Spinner
	case ast.SpinnerModeVisible:
		name := executionBlock.SpinnerName
		if name == "" {
			name = "Running..."
		}

		spinner := modifiers.NewSpinnerConstant(executionBlock.SpinnerName)
		intent.AddModifier(spinner)

		spinnerControl = spinner.Spinner
	}

	if executionBlock.ShowStdout {
		rutil.Logger.Trace().Msg("Streaming STDOUT")
		intent.AddModifier(modifiers.NewStdoutStream(writer))

		go func() {
			rutil.Logger.Trace().Msg("Setting up output formatter")

			prefix := termenv.String("  ")
			rutil.ReadAndFormatOutput(reader, 1, prefix.String(), spinnerControl /*bufio.NewWriter(r.Context.Output)*/, bufio.NewWriter(w), nil, "Running...")
		}()
	}

	if executionBlock.CaptureEnvironment {
		envCapture := modifiers.NewEnvironmentCapture()
		intent.AddModifier(envCapture)
	}

	result, err := intent.Execute()
	writer.Close()

	if err != nil {
		return goldast.WalkStop, err
	}

	if result.Env != nil {
		rutil.Logger.Debug().Msgf("Received environment: %#v", result.Env)

		r.Context.ImportEnv(result.Env)
	}

	w.Write(result.Output)

	return goldast.WalkContinue, nil
}
