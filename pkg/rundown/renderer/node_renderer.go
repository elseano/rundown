package renderer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/exec/modifiers"
	"github.com/elseano/rundown/pkg/exec/rpc"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/text"
	rutil "github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type RundownNodeRenderer struct {
	Context *Context
}

func NewRundownNodeRenderer(context *Context) *RundownNodeRenderer {
	return &RundownNodeRenderer{Context: context}
}

func (r *RundownNodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
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

func (r *RundownNodeRenderer) renderNoop(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	return goldast.WalkContinue, nil
}

func (r *RundownNodeRenderer) renderEnvironmentSubstitution(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
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

func (r *RundownNodeRenderer) renderString(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if str, ok := node.(*goldast.String); ok {
		if entering {
			w.Write(str.Value)
		}
	}

	return goldast.WalkContinue, nil
}

func (r *RundownNodeRenderer) renderTodo(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		w.WriteString(fmt.Sprintf("TODO - %s", node.Kind().String()))
	}

	return goldast.WalkContinue, nil
}

type BusSpinnerInterface struct {
	IsActive bool
}

func (s *BusSpinnerInterface) Start() {
	s.IsActive = true
	bus.Emit(&rpc.RpcMessage{Data: "STARTSPINNER"})
}

func (s *BusSpinnerInterface) Stop() {
	s.IsActive = false
	bus.Emit(&rpc.RpcMessage{Data: "STOPSPINNER"})
}

func (s *BusSpinnerInterface) Active() bool {
	return s.IsActive
}

func (r *RundownNodeRenderer) renderExecutionBlock(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
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

	intent, err := exec.NewExecution(executionBlock.With, script)
	if err != nil {
		return goldast.WalkStop, err
	}

	// intent.AddModifier(modifiers.NewStdout())
	reader, writer, _ := os.Pipe()

	if executionBlock.ShowStdout {
		rutil.Logger.Trace().Msg("Streaming STDOUT")
		intent.AddModifier(modifiers.NewStdoutStream(writer))

		spinner := BusSpinnerInterface{IsActive: true}

		go func() {
			rutil.Logger.Trace().Msg("Setting up output formatter")

			p := termenv.ColorProfile()

			prefix := termenv.String("‣ ")
			prefix.Foreground(p.Color("#5555ff"))
			rutil.ReadAndFormatOutput(reader, 1, prefix.String(), &spinner, bufio.NewWriter(w), nil, "Running...")
		}()
	}

	rutil.Logger.Debug().Msgf("Spinner mode %d", executionBlock.SpinnerMode)

	switch executionBlock.SpinnerMode {
	case ast.SpinnerModeVisible:
		spinner := modifiers.NewSpinnerConstant(executionBlock.SpinnerName)
		intent.AddModifier(spinner)
	}

	if executionBlock.CaptureEnvironment {
		envCapture := modifiers.NewEnvironmentCapture()
		intent.AddModifier(envCapture)
	}

	intent.ImportEnv(r.Context.Env)

	result, err := intent.Execute()
	writer.Close()

	if err != nil {
		return goldast.WalkStop, err
	}

	if result.Env != nil {
		r.Context.ImportEnv(result.Env)
	}

	w.Write(result.Output)

	return goldast.WalkContinue, nil
}
