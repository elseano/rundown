package segments

import (
	"bytes"
	"fmt"
	"log"
	"io"
	"math"
	"strings"

	"github.com/go-errors/errors"

	"github.com/elseano/rundown/markdown"
	"github.com/elseano/rundown/util"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

func BuildSegments(contents string, md goldmark.Markdown, logger *log.Logger) []Segment {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(errors.Wrap(err, 2).ErrorStack())
		}
	}()

	return doBuildSegments(contents, md, logger, 0)
}

func doBuildSegments(contents string, md goldmark.Markdown, logger *log.Logger, indent int) []Segment {
	reader := text.NewReader([]byte(contents))
	doc := md.Parser().Parse(reader)

	logger.Printf("Loaded file with %d block elements\n", doc.ChildCount())

	return PrepareSegments([]byte(contents), doc, logger, indent)
}

type SegmentType int
type ExecutionResult struct {
	Message string
	Kind string
	Source string
	Output string
	IsError bool
}

var (
	SuccessfulExecution = ExecutionResult{ Kind: "Success", IsError: false }
	SkipToNextHeading = ExecutionResult{ Kind: "Skip", IsError: false }
	AbortedExecution = ExecutionResult{ Kind: "Aborted", IsError: true }
)

const (
	DisplayBlock = iota + 1
	ExecuteBlock
)

type Segment interface {
	fmt.Stringer
	Execute(c *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) ExecutionResult
	Kind() string
	GetIndent() int
	DeIndent(amount int)
}

type BaseSegment struct {
	Segment
	Indent int
	Nodes  []ast.Node
	Source *[]byte
}

func (c *BaseSegment) DeIndent(amount int) {
	c.Indent = c.Indent - amount
	for _, node := range c.Nodes {
		if heading, ok := node.(*ast.Heading); ok {
			heading.Level = heading.Level - amount

			if heading.Level < 0 {
				heading.Level = 0
			}
		}
	}
}

func (c *BaseSegment) Execute(ctx *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	width, _, err := terminal.GetSize(0)

	if err != nil {
		width = 120
	}

	www := util.NewWordWrapWriter(out, int(math.Min(float64(width), 120)))

	www.SetAfterWrap(func(out io.Writer) int {
		n, _ := out.Write(bytes.Repeat([]byte("  "), c.Indent - 1))
		return n
	})

	for _, node := range c.Nodes {
		renderer.Render(www, *c.Source, node)
	}
	return SuccessfulExecution
}

func (s *BaseSegment) String() string {
	var buf bytes.Buffer

	buf.WriteString(s.Kind() + " {\n")
	body := util.CaptureStdout(func() {
		for _, n := range s.Nodes {
			n.Dump(*s.Source, s.Indent)
		}
	})
	buf.WriteString(body)
	buf.WriteString("}\n")

	return buf.String()
}

func (c *BaseSegment) Kind() string { return "Base" }

func (c *BaseSegment) GetIndent() int { return c.Indent }


type Separator struct {
	BaseSegment
}

func (c *Separator) Kind() string { return "Separator" }

func (c *Separator) Execute(ctx *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	out.Write([]byte("\r\n"))
	return SuccessfulExecution
}

type SetupSegment struct {
	BaseSegment
	Segment *CodeSegment
	HasRun bool
}

func (c *SetupSegment) Kind() string { return "SetupSegment" }

func (c *SetupSegment) Execute(ctx *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) ExecutionResult {
	if !c.HasRun {
		c.HasRun = true
		return c.Segment.Execute(ctx, renderer, lastSegment, logger, out)
	} else {
		return SuccessfulExecution
	}
}



type HeadingMarker struct {
	BaseSegment
	Title string
	ShortCode string
	Setup []*SetupSegment
	ParentHeading *HeadingMarker
}

func (c *HeadingMarker) Kind() string { return "HeadingMarker" }

func (c *HeadingMarker) RunSetups(ctx *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) (ExecutionResult, int) {
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

func (c *HeadingMarker) DeIndent(amount int) {
	c.BaseSegment.DeIndent(amount)

	if c.ParentHeading != nil {
		c.ParentHeading.DeIndent(amount)
	}

	// for _, node := range c.Setup {
	// 	node.DeIndent(amount)
	// }
}


type DisplaySegment struct {
	BaseSegment
}

func (c *DisplaySegment) Kind() string { return "DisplaySegment" }

func captureLines(v ast.Node, source []byte) string {
	var result = ""

	for i := 0; i < v.Lines().Len(); i++ {
		line := v.Lines().At(i)
		result = result + string(line.Value(source))
	}

	return result
}

func PrepareSegments(source []byte, doc ast.Node, logger *log.Logger, indent int) []Segment {
	var currentHeading *HeadingMarker
	var currentHeadingTree = map[int]*HeadingMarker{}
	var currentIndent = indent
	var currentContentSegment = &DisplaySegment{
		BaseSegment{
			Indent: 0,
			Nodes:  []ast.Node{},
			Source: &source,
		},
	}

	var segments = []Segment{currentContentSegment}

	logger.Printf("Constructing rundown segments")

	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if node.Kind() == ast.KindHeading {
			headingNode := node.(*ast.Heading)
			currentIndent = headingNode.Level
			lastSegment := segments[len(segments) - 1]

			if lastSegment.Kind() == "CodeSegment" {
				segments = append(segments, &Separator{
					BaseSegment {
						Indent: currentIndent,
						Nodes:  []ast.Node{},
						Source: &source,
					},
				})
			}

			var shortcode string = ""

			if lastSegment.Kind() == "DisplaySegment" {
				nodes := lastSegment.(*DisplaySegment).Nodes
				if len(nodes) > 0 {
					lastNode := nodes[len(nodes)-1]

					if lastNode.Kind() == markdown.KindCodeModifierBlock {
						cm := lastNode.(*markdown.CodeModifierBlock)
						mods := ParseModifiers(cm.Modifiers)

						if label, ok := mods.Values[LabelParameter]; ok {
							shortcode = label
						}
					}
				}
			}

			var parentHeading *HeadingMarker = nil
			if parent, ok := currentHeadingTree[currentIndent-1]; ok {
				parentHeading = parent
			}

			currentHeading = &HeadingMarker{
				BaseSegment: BaseSegment{
					Nodes:  []ast.Node{node},
					Indent: currentIndent,
					Source: &source,
				},
				Title: string(headingNode.Text(source)),
				ShortCode: shortcode,
				Setup: []*SetupSegment{},
				ParentHeading: parentHeading,
			}

			currentHeadingTree[currentIndent] = currentHeading

			segments = append(segments, currentHeading)

			currentContentSegment = nil
		} else if node.Kind() == ast.KindFencedCodeBlock {
			var mods = NewModifiers()

			nodes := []ast.Node{node}
			prev := node.PreviousSibling()
			if prev != nil && prev.Kind() == markdown.KindCodeModifierBlock {
				mods = ParseModifiers(prev.(*markdown.CodeModifierBlock).Modifiers)
			}

			info := node.(*ast.FencedCodeBlock).Info
			var infoText string

			if info != nil {
				infoText = string(info.Text(source))
			} else {
				infoText = ""
			}

			fencedMods := ParseModifiers(strings.Join(strings.Split(infoText, " ")[1:], " "))
			mods.Ingest(fencedMods)

			codeSegment := &CodeSegment{
				BaseSegment: BaseSegment{
					Indent: currentIndent,
					Nodes:  nodes,
					Source: &source,
				},
				Modifiers: mods,
				code:     captureLines(node, source),
				language: string(node.(*ast.FencedCodeBlock).Language(source)),
			}

			// Setup modifiers get attached to their headings. When we run headings, we ensure
			// they've executed all parent setups.
			if mods.Flags[SetupFlag] == true {
				setupSegment := &SetupSegment{ 
					BaseSegment: BaseSegment{
						Indent: currentIndent,
						Nodes:  nodes,
						Source: &source,	
					},
					Segment: codeSegment,
				}
				currentHeading.Setup = append(currentHeading.Setup, setupSegment)
				segments = append(segments, setupSegment)
			} else {
				segments = append(segments, codeSegment)
			}

			currentContentSegment = nil

		} else if node.Kind() == markdown.KindInvisibleBlock {
			invisible := node.(*markdown.InvisibleBlock)
			subDoc := captureLines(node, source)
			for _, subSegment := range doBuildSegments(subDoc, invisible.Markdown, logger, currentIndent) {
				segments = append(segments, subSegment)
			}
			currentContentSegment = nil
		} else {
			if currentContentSegment == nil {
				if segments[len(segments) - 1].Kind() == "CodeSegment" {
					segments = append(segments, &Separator{
						BaseSegment {
							Indent: currentIndent,
							Nodes:  []ast.Node{},
							Source: &source,
						},
					})
				}
	
				currentContentSegment = &DisplaySegment{
					BaseSegment{
						Indent: currentIndent,
						Nodes:  []ast.Node{},
						Source: &source,
					},
				}

				segments = append(segments, currentContentSegment)

			}
			currentContentSegment.Nodes = append(currentContentSegment.Nodes, node)
		}

	}

	return padSegments(segments)

	logger.Printf("Segments generated\n")

	return segments
}

func padSegments(segments []Segment) []Segment {
	return segments
}