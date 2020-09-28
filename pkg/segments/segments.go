package segments

import (
	"io"
	"log"

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

func (c *Separator) Execute(ctx *Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	out.Write([]byte("\r\n"))
	return SuccessfulExecution
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

func (c *SetupSegment) Execute(ctx *Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	if !c.HasRun {
		c.HasRun = true
		return c.Segment.Execute(ctx, renderer, lastSegment, logger, out)
	} else {
		return SuccessfulExecution
	}
}

type HeadingMarker struct {
	BaseSegment
	Title         string
	ShortCode     string
	Description   string
	Setup         []*SetupSegment
	ParentHeading *HeadingMarker
}

func (s *HeadingMarker) String() string {
	return s.Stringify("HeadingMarker", "")
}

func (c *HeadingMarker) Kind() string { return "HeadingMarker" }

func (c *HeadingMarker) RunSetups(ctx *Context, renderer renderer.Renderer, lastSegment Segment, logger *log.Logger, out io.Writer) (ExecutionResult, int) {
	var result ExecutionResult
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

	return SuccessfulExecution, count
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
