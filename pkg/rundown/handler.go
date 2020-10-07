package rundown

import (
	"errors"
	"io"
	"os"

	"github.com/yuin/goldmark/ast"

	"github.com/elseano/rundown/pkg/markdown"
)

type rundownHandler struct {
	// markdown.RundownHandler
	ctx *Context
}

func NewRundownHandler(ctx *Context) *rundownHandler {
	return &rundownHandler{
		ctx: ctx,
	}
}

func (v *rundownHandler) Mutate(input []byte, node ast.Node) ([]byte, error) {
	rundown := node.(markdown.RundownNode)

	if rundown.GetModifiers().Flags[EnvAwareFlag] {
		result, err := SubEnv(string(input), v.ctx)

		if err != nil {
			return input, err
		}

		return []byte(result), nil
	}

	return input, nil
}

func (v *rundownHandler) OnRundownNode(node ast.Node, entering bool) error {
	if !entering {
		if rundown, ok := node.(*markdown.RundownBlock); ok {
			if rundown.GetModifiers().Flags[StopOkFlag] {
				return &StopError{Result: StopOkResult}
			}

			if rundown.GetModifiers().Flags[StopFailFlag] {
				return &StopError{Result: StopFailResult}
			}
		}
	}

	return nil
}

func (v *rundownHandler) OnExecute(node *markdown.ExecutionBlock, source []byte, out io.Writer) (markdown.ExecutionResult, error) {
	result := Execute(v.ctx, node, source, v.ctx.Logger, os.Stdout)
	switch result {
	case SuccessfulExecution:
		return markdown.Continue, nil
	case SkipToNextHeading:

		nextNode := node.NextSibling()

		deleteNodesUntilSection(nextNode)

		return markdown.Continue, nil
	case StopOkResult:
		return markdown.Stop, nil
	case StopFailResult:
		return markdown.Stop, errors.New("Stop requested")
	}

	return markdown.Stop, &StopError{Result: result}
}

func deleteNodesUntilSection(node ast.Node) {
	for node != nil {
		if _, isSection := node.(*markdown.Section); isSection {
			return
		}

		nextNode := node.NextSibling()
		if nextNode == nil {
			nextNode = node.Parent().NextSibling()
		}

		node.Parent().RemoveChild(node.Parent(), node)

		node = nextNode
	}
}
