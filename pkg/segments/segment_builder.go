package segments

import (
	"container/list"
	"fmt"
	"log"

	"github.com/go-errors/errors"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type SegmentList struct {
	list               *list.List
	currentLevel       int
	currentHeading     *HeadingMarker
	source             *[]byte
	currentHeadingTree map[int]*HeadingMarker
}

func NewSegmentList(source *[]byte) *SegmentList {
	return &SegmentList{
		list:               list.New(),
		currentLevel:       1,
		currentHeading:     nil,
		source:             source,
		currentHeadingTree: map[int]*HeadingMarker{},
	}
}

func (s *SegmentList) List() *list.List {
	return s.list
}

func (s *SegmentList) LastSegment() Segment {
	back := s.list.Back()

	if back == nil {
		return nil
	}

	if s, ok := back.Value.(Segment); ok {
		return s
	}

	return nil
}

func (s *SegmentList) CurrentHeading() *HeadingMarker {
	return s.currentHeading
}

func (s *SegmentList) CurrentDisplaySegment() *DisplaySegment {
	if last := s.LastSegment(); last != nil && last.Kind() == "DisplaySegment" {
		return last.(*DisplaySegment)
	}

	level := 1
	if s.CurrentHeading() != nil {
		level = s.CurrentHeading().Level
	}

	display := &DisplaySegment{
		BaseSegment{
			Level:  level,
			Nodes:  []ast.Node{},
			Source: s.source,
		},
	}

	s.list.PushBack(display)

	return display

}

func (s *SegmentList) AppendAllContent(nodes []ast.Node) *DisplaySegment {
	display := s.CurrentDisplaySegment()

	for _, node := range nodes {
		display.AppendNode(node)
	}

	return display
}

func (s *SegmentList) AppendContent(node ast.Node) *DisplaySegment {
	display := s.CurrentDisplaySegment()
	display.AppendNode(node)

	return display
}

func (s *SegmentList) CurrentLevel() int {
	level := 1
	heading := s.CurrentHeading()
	if heading != nil {
		level = heading.Level
	}

	return level
}

func (s *SegmentList) AppendSegment(segment Segment) {
	s.list.PushBack(segment)

	if h, ok := segment.(*HeadingMarker); ok {
		s.currentHeading = h
	}
}

func (s *SegmentList) AppendCode(fcd *ast.FencedCodeBlock) {
	var mods = markdown.NewModifiers()
	if rundown, ok := fcd.PreviousSibling().(*markdown.RundownBlock); ok && rundown.ForCodeBlock {
		mods.Ingest(rundown.Modifiers)
	}

	var segment Segment
	segment = &CodeSegment{
		BaseSegment: BaseSegment{
			Level:  s.CurrentLevel(),
			Nodes:  []ast.Node{fcd},
			Source: s.source,
		},
		code:      util.NodeLines(fcd, *s.source),
		language:  string(fcd.Language(*s.source)),
		modifiers: mods,
	}

	if mods.Flags[rundown.SetupFlag] {
		segment = &SetupSegment{
			BaseSegment: BaseSegment{
				Level:  s.CurrentLevel(),
				Nodes:  []ast.Node{fcd},
				Source: s.source,
			},
			Segment: segment.(*CodeSegment),
		}

		s.CurrentHeading().AppendSetup(segment.(*SetupSegment))
	}

	s.AppendSegment(segment)
}

func (s *SegmentList) NewHeadingMarker(node *ast.Heading) *HeadingMarker {
	var shortcode = ""
	parent, hasParent := s.currentHeadingTree[node.Level-1]

	if !hasParent {
		parent = nil
	}

	if rundown, label := findRundownChildWithParameter(node, rundown.LabelParameter); rundown != nil {
		shortcode = label
	}

	marker := NewHeadingMarker(node, *s.source, parent)
	marker.ShortCode = shortcode
	s.AppendSegment(marker)

	s.currentHeadingTree[node.Level] = marker

	return marker
}

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

	logger.Printf("Constructing rundown segments")

	var list = NewSegmentList(&source)

	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if node.Kind() == ast.KindHeading {
			list.NewHeadingMarker(node.(*ast.Heading))

		} else if node.Kind() == ast.KindFencedCodeBlock {
			list.AppendCode(node.(*ast.FencedCodeBlock))

		} else if rd, ok := node.(*markdown.RundownBlock); ok {
			if rd.Modifiers.Flags[rundown.OnFailureFlag] {
				list.CurrentHeading().AppendHandler(rd)
			} else if _, ok := rd.Modifiers.Values[rundown.OnFailureParameter]; ok {
				list.CurrentHeading().AppendHandler(rd)
			} else if !rd.ForCodeBlock {
				list.AppendContent(rd)
			}
		} else {
			list.AppendContent(node)
		}

	}

	return padSegments(list)
}

func padSegments(segments *SegmentList) []Segment {
	var lastSegment Segment = nil
	var result = []Segment{}

	for element := segments.List().Front(); element != nil; element = element.Next() {
		segment := element.Value.(Segment)

		lastCode, lastIsCode := lastSegment.(*CodeSegment)
		lastSetup, lastIsSetup := lastSegment.(*SetupSegment)
		// lastDisplay, lastIsDisplay := lastSegment.(*DisplaySegment)
		currentHeading, currentIsHeading := segment.(*HeadingMarker)
		currentDisplay, currentIsDisplay := segment.(*DisplaySegment)
		currentCode, currentIsCode := segment.(*CodeSegment)

		if (lastIsCode || lastIsSetup) && currentIsHeading {
			result = append(result, NewSeparator(currentHeading.Level))
		}

		if lastIsCode && currentIsCode {
			if currentCode.modifiers.Flags[rundown.RevealFlag] && !lastCode.modifiers.Flags[rundown.RevealFlag] {
				result = append(result, NewSeparator(currentCode.Level))
			}
		}

		if lastIsSetup && currentIsCode {
			if currentCode.modifiers.Flags[rundown.RevealFlag] && !lastSetup.Segment.modifiers.Flags[rundown.RevealFlag] {
				result = append(result, NewSeparator(currentCode.Level))
			}
		}

		if (lastIsCode || lastIsSetup) && currentIsDisplay {
			result = append(result, NewSeparator(currentDisplay.Level))
		}

		result = append(result, segment)

		lastSegment = segment
	}

	return result
}
