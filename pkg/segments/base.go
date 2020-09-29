package segments

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type ExecutionResult struct {
	Message string
	Kind    string
	Source  string
	Output  string
	IsError bool
}

var (
	SuccessfulExecution = ExecutionResult{Kind: "Success", IsError: false}
	SkipToNextHeading   = ExecutionResult{Kind: "Skip", IsError: false}
	StopFailResult      = ExecutionResult{Kind: "Stop", IsError: true}
	StopOkResult        = ExecutionResult{Kind: "Stop", IsError: false}
)

type StopError struct {
	Result ExecutionResult
}

func (e *StopError) Error() string {
	return e.Result.Message
}

type Segment interface {
	fmt.Stringer
	Execute(c *Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) ExecutionResult
	Kind() string
	GetLevel() int
	DeLevel(amount int)
	GetModifiers() *markdown.Modifiers

	LastNode() ast.Node
	AppendNode(node ast.Node)
}

type RundownHandler struct {
	markdown.RundownHandler
	ctx *Context
}

func (v *RundownHandler) Mutate(input []byte, node ast.Node) ([]byte, error) {
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

func (v *RundownHandler) OnRundownNode(node ast.Node, entering bool) error {
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

func NewRundownHandler(ctx *Context) *RundownHandler {
	return &RundownHandler{
		ctx: ctx,
	}
}

type BaseSegment struct {
	Segment
	Level  int
	Nodes  []ast.Node
	Source *[]byte
}

func (c *BaseSegment) AppendNode(node ast.Node) {
	c.Nodes = append(c.Nodes, node)
}

func (c *BaseSegment) LastNode() ast.Node {
	if len(c.Nodes) > 0 {
		return c.Nodes[len(c.Nodes)-1]
	}

	return nil
}

func (c *BaseSegment) DeLevel(amount int) {
	c.Level = c.Level - amount
	for _, node := range c.Nodes {
		if heading, ok := node.(*ast.Heading); ok {
			heading.Level = heading.Level - amount

			if heading.Level < 0 {
				heading.Level = 0
			}
		}
	}
}

func renderNodes(renderer renderer.Renderer, nodes []ast.Node, source []byte, out io.Writer) error {
	doc := ast.NewDocument()

	for _, node := range nodes {
		doc.AppendChild(doc, node)
	}

	if err := renderer.Render(out, source, doc); err != nil {
		return err
	}

	return nil
}

func (c *BaseSegment) Execute(ctx *Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	// We can't change options on a renderer after it's rendered something
	// so we always use a fresh renderer.
	subRenderer := markdown.PrepareMarkdown().Renderer()

	currentLevel := c.Level

	subRenderer.AddOptions(markdown.WithRundownHandler(NewRundownHandler(ctx)))
	subRenderer.AddOptions(markdown.WithConsoleWidth(ctx.ConsoleWidth))
	subRenderer.AddOptions(markdown.WithLevelChange(func(indent int) {
		currentLevel = indent
	}))
	subRenderer.AddOptions(markdown.WithLevel(currentLevel))

	www := util.NewWordWrapWriter(out, ctx.ConsoleWidth)

	www.SetAfterWrap(func(out io.Writer) int {
		n, _ := out.Write(bytes.Repeat([]byte("  "), currentLevel-1))
		return n
	})

	if err := renderNodes(subRenderer, c.Nodes, *c.Source, www); err != nil {
		if er, ok := err.(*StopError); ok {
			return er.Result
		}

		return ExecutionResult{
			Message: err.Error(),
			Kind:    "Error",
			Source:  "",
			Output:  "",
			IsError: true,
		}
	}

	return SuccessfulExecution
}

func (s *BaseSegment) String() string {
	return s.Stringify("Base", "")
}

func (s *BaseSegment) Stringify(name string, extra string) string {
	var buf bytes.Buffer

	buf.WriteString(name + " {\n")
	if extra != "" {
		for _, line := range strings.Split(extra, "\n") {
			buf.WriteString("    " + line + "\n")
		}
	}

	buf.WriteString(fmt.Sprintf("Level: %d\n", s.Level))

	body := util.CaptureStdout(func() {
		for _, n := range s.Nodes {
			n.Dump(*s.Source, s.Level)
		}
	})
	buf.WriteString(body)
	buf.WriteString("}\n")

	return buf.String()
}

func (c *BaseSegment) Kind() string { return "Base" }

func (c *BaseSegment) GetLevel() int { return c.Level }

func (c *BaseSegment) GetModifiers() *markdown.Modifiers {
	return nil
}
