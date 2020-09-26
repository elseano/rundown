package segments

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-errors/errors"

	"github.com/elseano/rundown/markdown"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
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

const (
	DisplayBlock = iota + 1
	ExecuteBlock
)

func captureLines(v ast.Node, source []byte) string {
	var result = ""

	for i := 0; i < v.Lines().Len(); i++ {
		line := v.Lines().At(i)
		result = result + string(line.Value(source))
	}

	return result
}

func findRundownParameter(n ast.Node, param markdown.Parameter) (markdown.RundownNode, string) {
	if rundown, ok := n.(markdown.RundownNode); ok {
		if param, ok := rundown.GetModifiers().Values[param]; ok {
			return rundown, param
		}
	}

	return nil, ""
}

func findRundownChildWithFlag(n ast.Node, flag markdown.Flag) markdown.RundownNode {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if rundown, ok := c.(markdown.RundownNode); ok && rundown.GetModifiers().Flags[flag] {
			return rundown
		}
	}

	return nil
}

func findRundownChildWithParameter(n ast.Node, param markdown.Parameter) (markdown.RundownNode, string) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if rundown, value := findRundownParameter(c, param); rundown != nil {
			return rundown, value
		}
	}

	return nil, ""
}

func PrepareSegments(source []byte, doc ast.Node, logger *log.Logger, indent int) []Segment {
	var currentHeading *HeadingMarker
	var currentHeadingTree = map[int]*HeadingMarker{}
	var currentLevel = indent
	var currentContentSegment = &DisplaySegment{
		BaseSegment{
			Level:  0,
			Nodes:  []ast.Node{},
			Source: &source,
		},
	}

	var segments = []Segment{currentContentSegment}
	var rundownBlocks = []ast.Node{}

	logger.Printf("Constructing rundown segments")

	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if node.Kind() == ast.KindHeading {
			// Move any currently captured rundown blocks into the previous segment, if there is one.
			if len(rundownBlocks) > 0 {
				if currentContentSegment == nil {
					currentContentSegment = &DisplaySegment{
						BaseSegment{
							Level:  currentLevel,
							Nodes:  []ast.Node{},
							Source: &source,
						},
					}
					segments = append(segments, currentContentSegment)
				}
				for _, rundown := range rundownBlocks {
					currentContentSegment.AppendNode(rundown)
				}
			}

			headingNode := node.(*ast.Heading)
			currentLevel = headingNode.Level
			lastSegment := segments[len(segments)-1]

			// We'd like some space between the heading and the code block.
			if lastSegment.Kind() == "CodeSegment" {
				segments = append(segments, NewSeparator(currentLevel))
			}

			var shortcode string = ""

			// Is the first child a rundown block? Might be a label.
			if rundown, ok := node.NextSibling().(*markdown.RundownBlock); ok {
				if label, ok := rundown.Modifiers.Values[LabelParameter]; ok {
					shortcode = label
				}
			}

			// Otherwise, is there a rundown label specified in the heading?
			if rundown, label := findRundownChildWithParameter(node, LabelParameter); rundown != nil {
				shortcode = label
			}

			var parentHeading *HeadingMarker = nil
			if parent, ok := currentHeadingTree[currentLevel-1]; ok {
				parentHeading = parent
			}

			currentHeading = &HeadingMarker{
				BaseSegment: BaseSegment{
					Nodes:  []ast.Node{node},
					Level:  currentLevel,
					Source: &source,
				},
				Title:         strings.TrimSpace(string(headingNode.Text(source))),
				ShortCode:     shortcode,
				Setup:         []*SetupSegment{},
				ParentHeading: parentHeading,
			}

			if desc, ok := node.NextSibling().(*ast.Paragraph); ok {
				if rundown := findRundownChildWithFlag(desc, DescriptionFlag); rundown != nil {
					currentHeading.Description = string(rundown.Text(source))
				}
			} else if rundown, desc := findRundownParameter(node.NextSibling(), DescriptionParameter); rundown != nil {
				currentHeading.Description = desc
			}

			currentHeadingTree[currentLevel] = currentHeading

			segments = append(segments, currentHeading)

			rundownBlocks = []ast.Node{}

			currentContentSegment = nil
		} else if node.Kind() == ast.KindFencedCodeBlock {
			var mods = markdown.NewModifiers()
			var nodes = []ast.Node{}

			for _, rundown := range rundownBlocks {
				nodes = append(nodes, rundown)
				mods.Ingest(rundown.(*markdown.RundownBlock).Modifiers)
			}

			nodes = append(nodes, node)
			rundownBlocks = []ast.Node{}

			codeSegment := &CodeSegment{
				BaseSegment: BaseSegment{
					Level:  currentLevel,
					Nodes:  nodes,
					Source: &source,
				},
				code:      captureLines(node, source),
				language:  string(node.(*ast.FencedCodeBlock).Language(source)),
				modifiers: mods,
			}

			lastSegment := segments[len(segments)-1]
			if lastCode, ok := lastSegment.(*CodeSegment); ok {
				if mods.Flags[RevealFlag] && !lastCode.GetModifiers().Flags[RevealFlag] {
					// Add some space.
					segments = append(segments, NewSeparator(currentLevel))
				}
			}

			// Setup modifiers get attached to their headings. When we run headings, we ensure
			// they've executed all parent setups.
			if mods.Flags[SetupFlag] == true {
				setupSegment := &SetupSegment{
					BaseSegment: BaseSegment{
						Level:  currentLevel,
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

		} else if node.Kind() == markdown.KindRundownBlock && node.NextSibling() != nil {
			rundownBlocks = append(rundownBlocks, node)
		} else {
			if currentContentSegment == nil {
				if segments[len(segments)-1].Kind() == "CodeSegment" {
					segments = append(segments, &Separator{
						BaseSegment{
							Level:  currentLevel,
							Nodes:  []ast.Node{},
							Source: &source,
						},
					})
				}

				currentContentSegment = &DisplaySegment{
					BaseSegment{
						Level:  currentLevel,
						Nodes:  []ast.Node{},
						Source: &source,
					},
				}
				segments = append(segments, currentContentSegment)

			}

			for _, rundown := range rundownBlocks {
				currentContentSegment.AppendNode(rundown)
			}

			currentContentSegment.AppendNode(node)
			rundownBlocks = []ast.Node{}
		}

	}

	return padSegments(segments)

	logger.Printf("Segments generated\n")

	return segments
}

func padSegments(segments []Segment) []Segment {
	return segments
}
