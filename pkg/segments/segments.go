package segments

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Separator struct {
	BaseSegment
}

func NewSeparator(indent int) *Separator {
	return &Separator{
		BaseSegment: BaseSegment{Level: indent},
	}
}

func (s *Separator) String() string {
	return s.Stringify("Separator", "")
}

func (c *Separator) Kind() string { return "Separator" }

func (c *Separator) Execute(ctx *rundown.Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) rundown.ExecutionResult {
	out.Write([]byte("\r\n"))
	return rundown.SuccessfulExecution
}

type SetupSegment struct {
	BaseSegment
	Segment *CodeSegment
	HasRun  bool
}

func NewSetupSegment(indent int, segment *CodeSegment) *SetupSegment {
	return &SetupSegment{
		BaseSegment: BaseSegment{Level: indent},
		Segment:     segment,
		HasRun:      false,
	}
}

func (s *SetupSegment) String() string {
	return s.Stringify("Setup", "")
}

func (c *SetupSegment) Kind() string { return "SetupSegment" }

func (c *SetupSegment) Execute(ctx *rundown.Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) rundown.ExecutionResult {
	if !c.HasRun {
		c.HasRun = true
		return c.Segment.Execute(ctx, renderer, lastSegment, logger, out)
	} else {
		return rundown.SuccessfulExecution
	}
}

type Handler struct {
	BaseSegment
	Mods *markdown.Modifiers
}

func (s *Handler) String() string {
	return s.Stringify("Handler", "")
}

func (c *Handler) Kind() string { return "HeadingMarker" }

func NewHandler(node ast.Node, source []byte, level int) *Handler {
	rundown := node.(markdown.RundownNode)

	return &Handler{
		BaseSegment: BaseSegment{
			Nodes:  []ast.Node{node},
			Level:  level,
			Source: &source,
		},
		Mods: rundown.GetModifiers(),
	}
}

type HeadingMarker struct {
	BaseSegment
	Title         string
	ShortCode     string
	Description   string
	Setup         []*SetupSegment
	ParentHeading *HeadingMarker
	Handlers      []*Handler
}

func NewHeadingMarker(node ast.Node, source []byte, parent *HeadingMarker) *HeadingMarker {
	headingNode := node.(*ast.Heading)
	currentLevel := headingNode.Level

	var shortcode string = ""

	// Is the first child a rundown block? Might be a label.
	if rd, ok := node.NextSibling().(*markdown.RundownBlock); ok {
		if label, ok := rd.Modifiers.Values[rundown.LabelParameter]; ok {
			shortcode = label
		}
	}

	// Otherwise, is there a rundown label specified in the heading?
	if rundown, label := findRundownChildWithParameter(node, rundown.LabelParameter); rundown != nil {
		shortcode = label
	}

	currentHeading := &HeadingMarker{
		BaseSegment: BaseSegment{
			Nodes:  []ast.Node{node},
			Level:  currentLevel,
			Source: &source,
		},
		Title:         strings.TrimSpace(string(headingNode.Text(source))),
		ShortCode:     shortcode,
		Setup:         []*SetupSegment{},
		ParentHeading: parent,
		Handlers:      []*Handler{},
	}

	if desc, ok := node.NextSibling().(*ast.Paragraph); ok {
		if rundown := findRundownChildWithFlag(desc, rundown.DescriptionFlag); rundown != nil {
			currentHeading.Description = string(rundown.Text(source))
		}
	} else if rundown, desc := findRundownParameter(node.NextSibling(), rundown.DescriptionParameter); rundown != nil {
		currentHeading.Description = desc
	}

	return currentHeading
}

func (s *HeadingMarker) String() string {
	handlers := ""
	for _, h := range s.Handlers {
		handlers += h.String() + "\n"
	}
	setups := ""
	for _, h := range s.Setup {
		setups += h.String() + "\n"
	}

	return s.Stringify("HeadingMarker", fmt.Sprintf("Handlers: {%s\n}\nSetups: {%s\n}", handlers, setups))
}

func (s *HeadingMarker) AppendHandler(node ast.Node) {
	s.Handlers = append(s.Handlers, NewHandler(node, *s.Source, s.Level))
}

func (s *HeadingMarker) AppendSetup(setup *SetupSegment) {
	s.Setup = append(s.Setup, setup)
}

func (c *HeadingMarker) Kind() string { return "HeadingMarker" }

func (c *HeadingMarker) RunSetups(ctx *rundown.Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) (rundown.ExecutionResult, int) {
	var result rundown.ExecutionResult
	var count = 0

	if c.ParentHeading != nil {
		parentCount := 0
		result, parentCount = c.ParentHeading.RunSetups(ctx, renderer, lastSegment, logger, out)
		count = count + parentCount
	}

	if result.IsError {
		return result, count
	}

	for _, setup := range c.Setup {
		result = setup.Execute(ctx, renderer, lastSegment, logger, out)
		count = count + 1
		if result.IsError {
			return result, count
		}
	}

	return rundown.SuccessfulExecution, count
}

func (c *HeadingMarker) RunHandlers(errorText string, ctx *rundown.Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) rundown.ExecutionResult {
	for _, h := range c.Handlers {
		if match, ok := h.Mods.Values[rundown.OnFailureParameter]; ok {
			r, cerr := regexp.Compile(match)
			if cerr == nil {
				if r.MatchString(errorText) {
					h.Execute(ctx, renderer, lastSegment, logger, out)
				}
			}
		} else {
			h.Execute(ctx, renderer, lastSegment, logger, out)
		}
	}

	return rundown.SuccessfulExecution
}

func (c *HeadingMarker) DeLevel(amount int) {
	c.BaseSegment.DeLevel(amount)

	if c.ParentHeading != nil {
		c.ParentHeading.DeLevel(amount)
	}

	// for _, node := range c.Setup {
	// 	node.DeLevel(amount)
	// }
}

type DisplaySegment struct {
	BaseSegment
}

func (s *DisplaySegment) String() string {
	return s.Stringify("DisplaySegment", "")
}

func (c *DisplaySegment) Kind() string { return "DisplaySegment" }
