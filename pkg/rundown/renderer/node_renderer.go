package renderer

import (
	"io/ioutil"

	"github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/exec/modifiers"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/text"
	rutil "github.com/elseano/rundown/pkg/util"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type RundownNodeRenderer struct{}

func NewRundownNodeRenderer() *RundownNodeRenderer {
	return &RundownNodeRenderer{}
}

func (r *RundownNodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindExecutionBlock, r.renderExecutionBlock)
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

	if executionBlock.ShowStdout {
		intent.AddModifier(modifiers.NewStdout())
	}

	rutil.Logger.Debug().Msgf("Spinner mode %d", executionBlock.SpinnerMode)

	switch executionBlock.SpinnerMode {
	case ast.SpinnerModeVisible:
		intent.AddModifier(modifiers.NewSpinnerConstant(executionBlock.SpinnerName))
	}

	result, err := intent.Execute()

	if err != nil {
		return goldast.WalkStop, err
	}

	w.Write(result.Output)

	return goldast.WalkContinue, nil
}
