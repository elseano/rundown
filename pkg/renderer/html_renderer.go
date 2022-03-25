package renderer

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/exec/modifiers"
	"github.com/elseano/rundown/pkg/text"
	rutil "github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type RundownHtmlRenderer struct {
	Context *Context
}

func NewRundownHtmlRenderer(context *Context) *RundownHtmlRenderer {
	return &RundownHtmlRenderer{Context: context}
}

func (r *RundownHtmlRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(goldast.KindDocument, r.renderDocument)

	reg.Register(ast.KindExecutionBlock, r.renderExecutionBlock)
	reg.Register(ast.KindDescriptionBlock, r.renderNoop)
	reg.Register(ast.KindEnvironmentSubstitution, r.renderEnvironmentSubstitution)
	reg.Register(ast.KindIgnoreBlock, r.renderTodo)
	reg.Register(ast.KindOnFailure, r.renderTodo)
	reg.Register(ast.KindRundownBlock, r.renderTodo)
	reg.Register(ast.KindSaveCodeBlock, r.renderTodo)
	reg.Register(ast.KindSectionEnd, r.renderNoop)
	reg.Register(ast.KindSectionOption, r.renderOptionInput)
	reg.Register(ast.KindSectionPointer, r.renderNoop)
	reg.Register(ast.KindStopFail, r.renderStop)
	reg.Register(ast.KindStopOk, r.renderStop)

	reg.Register(goldast.KindString, r.renderString)
}

func (r *RundownHtmlRenderer) renderDocument(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		w.WriteString("<html><head>")
		w.WriteString("<style>")
		w.WriteString(css())
		w.WriteString("</style>")
		w.WriteString("</head><body class='container'><main class='content'>")
		w.WriteString("<form method='POST'>")
	} else {
		w.WriteString("<div class=\"field\"><div class=\"control\"><button class=\"button is-link\">Run</button></div></div></form>")
		w.WriteString("</main></body>")
	}

	return goldast.WalkContinue, nil
}

func (r *RundownHtmlRenderer) renderOptionInput(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		optNode := node.(*ast.SectionOption)

		w.WriteString("<div class='field is-horizontal'>")
		w.WriteString("<div class='field-label is-normal'>")
		w.WriteString("<label class='label'>" + optNode.OptionName + "</label>")
		w.WriteString("</div><div class='field-body'><div class='field'>")
		w.WriteString("<div class='control'>")
		w.WriteString("<input type='text' class='input' name='" + optNode.OptionName + "'/>")
		w.WriteString("</div>")
		w.WriteString("<p class='help'>" + optNode.OptionDescription + " (" + optNode.OptionType.Describe() + ")</p>")
		w.WriteString("</div></div>")
		w.WriteString("</div>")
	}

	return goldast.WalkContinue, nil
}

func (r *RundownHtmlRenderer) renderNoop(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	return goldast.WalkContinue, nil
}

func (r *RundownHtmlRenderer) renderEnvironmentSubstitution(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
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

func (r *RundownHtmlRenderer) renderString(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if str, ok := node.(*goldast.String); ok {
		if entering {
			w.Write(str.Value)
		}
	}

	return goldast.WalkContinue, nil
}

func (r *RundownHtmlRenderer) renderStop(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		switch node.(type) {
		case *ast.StopOk:
			io.WriteString(w, "<div class='info info-success'>")
		case *ast.StopFail:
			io.WriteString(w, "<div class='info info-error'>")
		}

		return goldast.WalkContinue, nil
	} else {
		io.WriteString(w, "</div>")

		return goldast.WalkStop, nil
	}
}

func (r *RundownHtmlRenderer) renderTodo(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	if entering {
		w.WriteString(fmt.Sprintf("TODO - %s", node.Kind().String()))
	}

	return goldast.WalkContinue, nil
}

func updateSpinner(w util.BufWriter, id string, title string, status string) {
	io.WriteString(w, "<script>document.getElementById('status_"+id+"').innerHTML = \""+status+"\";")
	io.WriteString(w, "document.getElementById('title_"+id+"').innerHTML = \""+title+"\";")
	io.WriteString(w, "</script>")
}

func (r *RundownHtmlRenderer) renderExecutionBlock(w util.BufWriter, source []byte, node goldast.Node, entering bool) (goldast.WalkStatus, error) {
	executionBlock := node.(*ast.ExecutionBlock)
	if !executionBlock.Execute {
		return goldast.WalkContinue, nil
	}

	if entering {
		io.WriteString(w, "<div class='columns'><div class='column'><span id='status_"+executionBlock.ID+"' class='tag is-info'>R</span> <span id='title_"+executionBlock.ID+"'>Running...</span></div><div class='column is-two-thirds'><progress class='progress' id='"+executionBlock.ID+"' max='100%'></progress></div></div>")
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

	if executionBlock.ReplaceProcess {
		intent.ReplaceProcess = true
		_, err := intent.Execute()

		if err != nil {
			return goldast.WalkStop, err
		}
	}

	rutil.Logger.Debug().Msgf("Spinner mode %d", executionBlock.SpinnerMode)

	var spinnerControl modifiers.SpinnerControl

	// switch executionBlock.SpinnerMode {
	// case ast.SpinnerModeInlineAll:
	// 	spinner := modifiers.NewSpinnerConstant(executionBlock.SpinnerName)
	// 	intent.AddModifier(spinner)

	// 	rutil.Logger.Debug().Msg("Inline all mode")
	// 	spinnerDetector := modifiers.NewSpinnerFromScript(true, spinner)
	// 	intent.AddModifier(spinnerDetector)

	// 	spinnerControl = spinner.Spinner
	// case ast.SpinnerModeVisible:
	// 	name := executionBlock.SpinnerName
	// 	if name == "" {
	// 		name = "Running..."
	// 	}

	// 	spinner := modifiers.NewSpinnerConstant(name)
	// 	intent.AddModifier(spinner)

	// 	updateSpinner(w, executionBlock.ID, name, "R")

	// 	spinnerControl = spinner.Spinner
	// }

	if executionBlock.SubstituteEnvironment {
		intent.AddModifier(modifiers.NewReplace(r.Context.Env))
	}

	doneChan := make(chan bool, 1)

	if executionBlock.ShowStdout {
		rutil.Logger.Trace().Msg("Streaming STDOUT")
		w.WriteString("<pre>")
		stdout := modifiers.NewStdoutStream()
		intent.AddModifier(stdout)

		go func() {
			rutil.Logger.Trace().Msg("Setting up output formatter")

			prefix := termenv.String("  ")
			rutil.ReadAndFormatOutput(stdout.Reader, 1, prefix.String(), spinnerControl /*bufio.NewWriter(r.Context.Output)*/, bufio.NewWriter(w), nil, "Running...")

			rutil.Logger.Trace().Msg("Output stream ended")

			w.WriteString("</pre>")
			doneChan <- true
		}()
	}

	if executionBlock.CaptureEnvironment != nil {
		envCapture := modifiers.NewEnvironmentCapture(executionBlock.CaptureEnvironment)
		intent.AddModifier(envCapture)
	}

	result, err := intent.Execute()
	rutil.Logger.Trace().Msgf("Execution complete: %v", result)

	if result.Env != nil {
		for _, name := range executionBlock.CaptureEnvironment {
			r.Context.ImportEnv(map[string]string{name: result.Env[name]})
		}
	}

	rutil.Logger.Trace().Msg("Waiting on done channel")

	<-doneChan

	w.Write(result.Output)
	w.WriteString("<script>document.getElementById('" + executionBlock.ID + "').remove()</script>")

	if err != nil {
		updateSpinner(w, executionBlock.ID, err.Error(), "ERR")
		return goldast.WalkStop, err
	} else if result.ExitCode != 0 {
		updateSpinner(w, executionBlock.ID, fmt.Sprintf("Failed with code %d", result.ExitCode), "ERR")
		return goldast.WalkStop, err
	}

	updateSpinner(w, executionBlock.ID, "Running... (Complete)", "OK")

	rutil.Logger.Trace().Msg("End render execution block")

	return goldast.WalkContinue, nil
}
