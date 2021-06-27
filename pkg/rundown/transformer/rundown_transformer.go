package transformer

import (
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/util"

	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	goldtext "github.com/yuin/goldmark/text"
)

// NodeProcessors allow us to apply effects to subsequent nodes or "child" nodes of a Rundown tag.
type NodeProcessor interface {
	Begin(openingTag *RundownHtmlTag)

	// Process a Markdown Node. Returns true to indicate the processor is done and should be removed.
	Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool

	// Process a closing Rundown Element. Returns true to indicate the processor is done and should be removed.
	End(node goldast.Node, openingTag *RundownHtmlTag, treatments *Treatment) bool
}

type rundownASTTransformer struct {
}

var defaultRundownASTTransformer = &rundownASTTransformer{}

// Rundown AST Transformer converts Rundown Elements in the markdown tree
// into proper rundown nodes, and applies any effects.
func NewRundownASTTransformer() parser.ASTTransformer {
	return defaultRundownASTTransformer
}

func (a *rundownASTTransformer) Transform(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var openNodes = []*RundownHtmlTag{}
	var activeProcessors = []NodeProcessor{}

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
						openingElement := openNodes[len(openNodes)-1]
						openNodes = openNodes[0 : len(openNodes)-1]

						newActiveProcessors := []NodeProcessor{}

						for i := 0; i < len(activeProcessors); i++ {
							if !activeProcessors[i].End(node, openingElement, treatments) {
								newActiveProcessors = append(newActiveProcessors, activeProcessors[i])
							}
						}

						activeProcessors = newActiveProcessors
					}

				} else {
					if processor != nil {
						processor.Begin(htmlNode)
						activeProcessors = append(activeProcessors, processor)
					}

					if !htmlNode.closed {
						openNodes = append(openNodes, htmlNode)
					}
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

			newActiveProcessors := []NodeProcessor{}

			for i := 0; i < len(activeProcessors); i++ {
				if !activeProcessors[i].Process(node, reader, treatments) {
					newActiveProcessors = append(newActiveProcessors, activeProcessors[i])
				}
			}

			activeProcessors = newActiveProcessors

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

				return &SectionProcessor{SectionPointer: section}
			} else {
				node.Parent().InsertBefore(node.Parent(), node, section)
				section.StartNode = node.NextSibling()

				return &SectionProcessor{SectionPointer: section}
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

	nextMeaningfulNode := node.NextSibling()
	if nextMeaningfulNode == nil {
		if para, ok := node.Parent().(*goldast.Paragraph); ok && para.ChildCount() == 1 {
			nextMeaningfulNode = para.NextSibling()
		}
	}

	if _, ok := nextMeaningfulNode.(*goldast.FencedCodeBlock); ok && data.HasAttr("with", "spinner", "stdout", "subenv", "sub-env", "capture-env", "replace", "borg", "reveal", "reveal-only") {
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

			if !executionBlock.Reveal {
				treatments.Remove(node)
			} else {
				node.Parent().InsertBefore(node.Parent(), nodeToReplace, node)
				treatments.Ignore(node)
			}

			return nil
		}
	}

	if data.HasAttr("subenv", "sub-env") {
		return &SubEnvProcessor{}
	}

	return nil
}
