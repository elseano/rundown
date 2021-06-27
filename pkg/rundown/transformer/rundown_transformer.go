package transformer

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/text"
	"github.com/elseano/rundown/pkg/util"

	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	goldtext "github.com/yuin/goldmark/text"
)

type rundownASTTransformer struct {
}

var defaultRundownASTTransformer = &rundownASTTransformer{}

// Rundown AST Transformer converts Rundown Elements in the markdown tree
// into proper rundown nodes.
func NewRundownASTTransformer() parser.ASTTransformer {
	return defaultRundownASTTransformer
}

type RundownHtmlTag struct {
	tag      string
	attrs    []html.Attribute
	contents string
	closed   bool
	closer   bool
}

func (r *RundownHtmlTag) HasAttr(names ...string) bool {
	for _, name := range names {
		for _, a := range r.attrs {
			if a.Key == name {
				return true
			}
		}
	}

	return false
}

func (r *RundownHtmlTag) GetAttr(name string) *string {
	for _, a := range r.attrs {
		if a.Key == name {
			return &a.Val
		}
	}

	return nil
}

func ExtractRundownElement(node goldast.Node, reader goldtext.Reader, currentTag string) *RundownHtmlTag {
	z := html.NewTokenizerFragment(text.NewNodeReader(node, reader), currentTag)
	// source := string(node.Text(reader.Source()))
	// z := html.NewTokenizer(strings.NewReader(source))

	var currentRundownTag *RundownHtmlTag

	for {
		ttype := z.Next()
		token := z.Token()
		switch ttype {
		case html.StartTagToken, html.SelfClosingTagToken:
			if token.Data == "r" || strings.HasPrefix(token.Data, "r-") {
				currentRundownTag = &RundownHtmlTag{
					tag:    token.Data,
					closed: false,
				}
			}

			if token.Attr != nil {
				currentRundownTag.attrs = make([]html.Attribute, len(token.Attr))
				copy(currentRundownTag.attrs, token.Attr)
			}

			if ttype == html.SelfClosingTagToken {
				currentRundownTag.closed = true
				return currentRundownTag
			}
		case html.TextToken:
			if currentRundownTag != nil {
				currentRundownTag.contents = currentRundownTag.contents + token.Data
			}
		case html.EndTagToken:
			if currentRundownTag != nil {
				currentRundownTag.closed = true
				return currentRundownTag
			}

			// Otherwise, we might have the closing tag of an earlier opened tag.
			if token.Data == "r" || strings.HasPrefix(token.Data, "r-") {
				return &RundownHtmlTag{
					tag:    token.Data,
					closer: true,
				}
			}

			return nil

		// ErrorToken is expected for inline RawHTML nodes, as they don't contain the entire HTML element,
		// instead there's a RawHTML for the opening, and a RawHTML for the closing tag.
		case html.ErrorToken:
			return currentRundownTag
		}

	}
}

type Treatment struct {
	replaceNodes []func()
	ignoreNodes  map[goldast.Node]bool
	reader       goldtext.Reader
}

func NewTreatment(reader goldtext.Reader) *Treatment {
	return &Treatment{
		replaceNodes: make([]func(), 0),
		ignoreNodes:  map[goldast.Node]bool{},
		reader:       reader,
	}
}

func (t *Treatment) Replace(nodeToReplace goldast.Node, replacement goldast.Node) {
	t.replaceNodes = append(t.replaceNodes, func() {
		parent := nodeToReplace.Parent()
		if parent == nil {
			return // Ignore, already removed.
		}

		if replacement.Parent() == nil {
			nodeToReplace.Parent().ReplaceChild(nodeToReplace.Parent(), nodeToReplace, replacement)
		}
	})
}

// Remove a node. Returns what the next node will be after this node is removed.
func (t *Treatment) Remove(nodeToRemove goldast.Node) goldast.Node {
	t.replaceNodes = append(t.replaceNodes, func() {
		// Handle node already removed
		if nodeToRemove.Parent() != nil {
			// Trim spacing between rundown element and the previous node.
			switch prev := nodeToRemove.PreviousSibling().(type) {
			case *goldast.Text:
				prev.Segment = prev.Segment.TrimRightSpace(t.reader.Source())
			}
			nodeToRemove.Parent().RemoveChild(nodeToRemove.Parent(), nodeToRemove)

		}
	})

	nextNode := nodeToRemove.NextSibling()

	if nodeToRemove.Parent().ChildCount() == 1 {
		if p, ok := nodeToRemove.Parent().(*goldast.Paragraph); ok {
			nextNode = p.NextSibling()
		}
	}

	if nextNode == nil {
		nextNode = nodeToRemove.Parent().NextSibling()
	}

	return nextNode
}

func (t *Treatment) AppendChild(parent goldast.Node, child goldast.Node) {
	t.replaceNodes = append(t.replaceNodes, func() {
		parent.AppendChild(parent, child)
	})
}

func (t *Treatment) Ignore(nodeToIgnore goldast.Node) {
	t.ignoreNodes[nodeToIgnore] = true
}

func (t *Treatment) IsIgnored(nodeInQuestion goldast.Node) bool {
	if ig, ok := t.ignoreNodes[nodeInQuestion]; ok {
		return ig
	}

	return false
}

func (t *Treatment) Process(reader goldtext.Reader) {
	for _, replacement := range t.replaceNodes {
		replacement()
	}

}

type NodeProcessor interface {
	Begin()
	Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment)
	End(treatments *Treatment)
}

type OpenElement struct {
	element   *RundownHtmlTag
	processor NodeProcessor
}

func (a *rundownASTTransformer) Transform(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var openNodes = []OpenElement{}

	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch node := node.(type) {

		case *goldast.RawHTML, *goldast.HTMLBlock:
			if htmlNode := ExtractRundownElement(node, reader, ""); htmlNode != nil {
				processor := ConvertToRundownNode(htmlNode, node, reader, treatments)

				util.Logger.Debug().Msgf("Got HTML node %#v", htmlNode)

				if htmlNode.closer {
					if len(openNodes) > 0 {
						openNodes[len(openNodes)-1].processor.End(treatments)
						openNodes = openNodes[0 : len(openNodes)-1]
					}

				} else if !htmlNode.closed {
					if processor != nil {
						processor.Begin()
					}

					openNodes = append(openNodes, OpenElement{element: htmlNode, processor: processor})
				}

				// Don't leave the Rundown Element in the document tree.
				treatments.Remove(node)
			}

		case *goldast.FencedCodeBlock:
			if !treatments.IsIgnored(node) {
				eb := ast.NewExecutionBlock(node)
				eb.With = string(node.Info.Text(reader.Source()))
				treatments.Replace(node, eb)
			}

		default:

			for _, open := range openNodes {
				if open.processor != nil {
					open.processor.Process(node, reader, treatments)
				}
			}

		}

		return goldast.WalkContinue, nil
	})

	treatments.Process(reader)

}

func ConvertToRundownNode(data *RundownHtmlTag, node goldast.Node, reader goldtext.Reader, treatments *Treatment) NodeProcessor {
	if data.HasAttr("label", "section") {
		// Label is the deprecated name.
		name := data.GetAttr("label")
		if name == nil {
			name = data.GetAttr("section")
		}

		if name != nil {
			util.Logger.Debug().Msgf("Section: %s", *name)

			// Marker can be in a paragraph, or in a heading. Either way, the containing node is the section starting node.
			section := ast.NewSectionPointer(*name)

			if heading, ok := node.Parent().(*goldast.Heading); ok {
				util.Logger.Debug().Msg("Parent is a heading")
				heading.Parent().InsertBefore(heading.Parent(), heading, section)
				section.StartNode = heading
			}

		}

	}

	if data.HasAttr("save") {
		nextNode := node.NextSibling()
		nodeToReplace := node

		if para, ok := node.Parent().(*goldast.Paragraph); ok && para.ChildCount() == 1 {
			nodeToReplace = para
			nextNode = para.NextSibling()
		}

		if node, ok := nextNode.(*goldast.FencedCodeBlock); ok {
			saveName := *data.GetAttr("save")
			executionBlock := ast.NewSaveCodeBlock(node, saveName)

			treatments.Ignore(node)
			treatments.Replace(nodeToReplace, executionBlock)
		}

		return nil
	}

	if data.HasAttr("opt") {
		opt := ast.NewSectionOption(*data.GetAttr("opt"))
		opt.OptionRequired = data.HasAttr("required")
		opt.OptionPrompt = data.GetAttr("prompt")

		if data.HasAttr("type") {
			opt.OptionType = ast.BuildOptionType(*data.GetAttr("type"))
		} else {
			opt.OptionType = ast.BuildOptionType("string")
		}

		if data.HasAttr("default") {
			def := opt.OptionType.Normalise(*data.GetAttr("default"))
			opt.OptionDefault = &def
		}

		treatments.Replace(node, opt)
	}

	if data.HasAttr("desc") {
		descNode := ast.NewDescriptionBlock()

		if data.closed {
			descNode.AppendChild(descNode, goldast.NewString([]byte(*data.GetAttr("desc"))))
			treatments.Replace(node, descNode)
		} else {
			if _, ok := node.Parent().(*goldast.Paragraph); ok {
				return NewGatherProcessor(node.Parent(), descNode)
			}
		}
	}

	if data.HasAttr("stop-fail") {
		stop := ast.NewStopFail()

		if data.closed {
			stop.AppendChild(stop, goldast.NewString([]byte(*data.GetAttr("stop-fail"))))
			treatments.Replace(node, stop)
		} else {
			if _, ok := node.Parent().(*goldast.Paragraph); ok {
				return NewGatherProcessor(node.Parent(), stop)
			}
		}
	}

	if data.HasAttr("stop-ok") {
		stop := ast.NewStopFail()

		if data.closed {
			stop.AppendChild(stop, goldast.NewString([]byte(*data.GetAttr("stop-ok"))))
			treatments.Replace(node, stop)
		} else {
			if _, ok := node.Parent().(*goldast.Paragraph); ok {
				return NewGatherProcessor(node.Parent(), stop)
			}
		}
	}

	if data.HasAttr("ignore") {
		// Just delete everything until the closing tag.
		if !data.closed {
			if _, ok := node.Parent().(*goldast.Paragraph); ok {
				return NewGatherProcessor(node.Parent(), nil)
			}
		}
	}

	if data.HasAttr("on-failure") {
		fail := ast.NewOnFailure()

		fail.FailureMessageRegexp = *data.GetAttr("on-failure")

		if _, ok := node.Parent().(*goldast.Paragraph); ok {
			return NewGatherProcessor(node.Parent(), fail)
		}
	}

	if data.HasAttr("with", "spinner", "stdout", "subenv", "sub-env", "capture-env", "replace", "borg") {
		nextNode := node.NextSibling()
		nodeToReplace := node

		if para, ok := node.Parent().(*goldast.Paragraph); ok && para.ChildCount() == 1 {
			nodeToReplace = para
			nextNode = para.NextSibling()
		}

		if node, ok := nextNode.(*goldast.FencedCodeBlock); ok {
			executionBlock := ast.NewExecutionBlock(node)

			executionBlock.ShowStdout = data.HasAttr("stdout")
			executionBlock.ShowStderr = data.HasAttr("stderr")
			executionBlock.Reveal = data.HasAttr("reveal", "reveal-only")
			executionBlock.Execute = !data.HasAttr("reveal-only", "norun")
			executionBlock.SubstituteEnvironment = data.HasAttr("subenv") || data.HasAttr("sub-env")
			executionBlock.CaptureEnvironment = data.HasAttr("capture-env")
			executionBlock.ReplaceProcess = data.HasAttr("borg")

			if spinnerName := data.GetAttr("spinner"); spinnerName != nil {
				executionBlock.SpinnerName = *spinnerName
				executionBlock.SpinnerMode = ast.SpinnerModeVisible
			} else if data.HasAttr("nospin") {
				executionBlock.SpinnerMode = ast.SpinnerModeHidden
			} else if data.HasAttr("named") {
				executionBlock.SpinnerMode = ast.SpinnerModeInlineFirst
			} else if data.HasAttr("named-all") {
				executionBlock.SpinnerMode = ast.SpinnerModeInlineAll
			}

			if withVal := data.GetAttr("with"); withVal != nil {
				executionBlock.With = *withVal
			} else {
				executionBlock.With = string(node.Info.Text(reader.Source()))
			}

			treatments.Replace(nodeToReplace, executionBlock)
			treatments.Remove(node) // Ensure we're removing the FCB

			return nil
		}
	}

	if data.HasAttr("subenv", "sub-env") {
		return &SubEnvProcessor{}
	}

	return nil
}
