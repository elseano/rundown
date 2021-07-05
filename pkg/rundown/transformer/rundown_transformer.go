package transformer

import (
	"fmt"
	"strings"

	"github.com/elseano/rundown/pkg/rundown/ast"
	"gopkg.in/guregu/null.v4"

	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	goldtext "github.com/yuin/goldmark/text"
)

// NodeProcessors allow us to apply effects to subsequent nodes or "child" nodes of a Rundown tag.
type NodeProcessor interface {
	Begin(openingTag *ast.RundownBlock)

	// Process a Markdown Node. Returns true to indicate the processor is done and should be removed.
	Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool

	// Indicates to the processor it should end itself.
	End(node goldast.Node, reader goldtext.Reader, treatments *Treatment)
}

type rundownASTTransformer struct {
}

var defaultRundownASTTransformer = &rundownASTTransformer{}

// Rundown AST Transformer converts Rundown Elements in the markdown tree
// into proper rundown nodes, and applies any effects.
func NewRundownASTTransformer() parser.ASTTransformer {
	return defaultRundownASTTransformer
}

type OpenTags struct {
	data *RundownHtmlTag
	node goldast.Node
}

func createRundownBlocks(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var openNodes = []OpenTags{}

	doc.Dump(reader.Source(), 0)

	// First, transform rundown opening/closing RawHTML into RundownBlocks.
	// This makes the next phase simpler in terms of handling what's inside any rundown block forms.
	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch node := node.(type) {

		case *goldast.RawHTML, *goldast.HTMLBlock:
			if htmlNode := ExtractRundownElement(node, reader, ""); htmlNode != nil {

				if htmlNode.closed {
					// No content.
					rdb := ast.NewRundownBlock()
					rdb.Attrs = htmlNode.attrs
					rdb.TagName = htmlNode.tag

					treatments.Replace(node, rdb)
					treatments.Remove(node)
				} else if htmlNode.closer {
					// Create block.
					if len(openNodes) > 0 {
						openingElement := openNodes[len(openNodes)-1]
						openNodes = openNodes[0 : len(openNodes)-1]

						rdb := ast.NewRundownBlock()
						rdb.Attrs = openingElement.data.attrs
						rdb.TagName = openingElement.data.tag

						// Move all nodes between start and end into rdb.
						var nextChild goldast.Node
						for child := openingElement.node.NextSibling(); child != nil && child != node; child = nextChild {
							nextChild = child.NextSibling()
							rdb.AppendChild(rdb, child)
						}

						treatments.Replace(openingElement.node, rdb)
						treatments.Remove(node)
					}
				} else {
					openNodes = append(openNodes, OpenTags{data: htmlNode, node: node})
				}

			}
		}

		return goldast.WalkContinue, nil
	})

	treatments.Process(reader)

	fmt.Printf("Rundown Blocks:\n")
	doc.Dump(reader.Source(), 0)
}

func (a *rundownASTTransformer) Transform(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	createRundownBlocks(doc, reader, pc)
	convertRundownBlocks(doc, reader, pc)
}

func convertRundownBlocks(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var activeProcessors = []NodeProcessor{}

	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch node := node.(type) {

		case *ast.RundownBlock:
			processor := ConvertToRundownNode(node, reader, treatments)

			if processor != nil {
				processor.Begin(node)
				activeProcessors = append(activeProcessors, processor)
			}

			// Don't leave the Rundown Element in the document tree.
			// treatments.Remove(node)

		case *goldast.FencedCodeBlock:
			if !treatments.IsIgnored(node) {
				eb := ast.NewExecutionBlock(node)
				eb.With = string(node.Info.Text(reader.Source()))
				treatments.Replace(node, eb)
			}

		}

		nodesToProcess := append(treatments.NewNodes(), node)

		for _, ntp := range nodesToProcess {

			newActiveProcessors := []NodeProcessor{}

			for i := 0; i < len(activeProcessors); i++ {
				if !activeProcessors[i].Process(ntp, reader, treatments) {
					newActiveProcessors = append(newActiveProcessors, activeProcessors[i])
				}
			}

			activeProcessors = newActiveProcessors
		}

		return goldast.WalkContinue, nil
	})

	// Close out any active processors.
	for _, p := range activeProcessors {
		p.End(nil, reader, treatments)
	}

	fmt.Printf("Unprocessed doc is:\n")
	doc.Dump(reader.Source(), 2)
	treatments.Process(reader)
	fmt.Printf("Processed doc is:\n")
	doc.Dump(reader.Source(), 2)

	// Populate sections
	goldast.Walk(doc, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			if end, ok := n.(*ast.SectionEnd); ok {
				fmt.Printf("Found section end\n")
				section := end.SectionPointer
				PopulateSectionMetadata(section, end, reader)
			}
		}

		return goldast.WalkContinue, nil
	})

	fmt.Printf("Sections populated\n")
}

func ConvertToRundownNode(node *ast.RundownBlock, reader goldtext.Reader, treatments *Treatment) NodeProcessor {
	var nodeToReplace goldast.Node = node
	nextNode := node.NextSibling()
	parentNode := nodeToReplace.Parent()

	// If we're the only child of a paragraph, then this is a stand-alone rundown tag. Replace the paragraph instead of the node
	// when inserting the rundown tag into the document.
	if para, ok := nodeToReplace.Parent().(*goldast.Paragraph); ok {
		if para.ChildCount() == 1 {
			nodeToReplace = para
			nextNode = para.NextSibling()
		} else {
			para.Dump(reader.Source(), 0)
		}
	}

	if node.HasAttr("label", "section") {
		name := node.GetAttr("section")
		if !name.Valid {
			name = node.GetAttr("label")
		}

		start := ast.NewSectionPointer(name.String)

		if name.Valid {
			if heading, ok := parentNode.(*goldast.Heading); ok {
				start.StartNode = heading
				start.DescriptionShort = strings.Trim(string(heading.Text(reader.Source())), " ")

				heading.Parent().InsertBefore(heading.Parent(), heading, start)

				end := FindEndOfSection(heading)
				if end != nil {
					end.Parent().InsertBefore(end.Parent(), end, &ast.SectionEnd{SectionPointer: start})
				} else {
					node.OwnerDocument().AppendChild(node.OwnerDocument(), &ast.SectionEnd{SectionPointer: start})
				}

				treatments.Remove(node)
			} else {
				// Section is not attached to a heading. Surround the Rundown Block with the section markers,
				// and dissolve the Rundown Block.
				start.StartNode = node.NextSibling()
				start.DescriptionShort = node.GetAttr("desc").ValueOrZero()

				node.Parent().InsertBefore(node.Parent(), node, start)

				end := &ast.SectionEnd{SectionPointer: start}
				node.Parent().InsertAfter(node.Parent(), node, end)

				treatments.DissolveRundownBlock(node)
			}
		}

		return nil
	}

	// if node.HasAttr("label", "section") {
	// 	// Label is the deprecated name.
	// 	name := node.GetAttr("section")
	// 	if name.Valid == false {
	// 		name = node.GetAttr("label")
	// 	}

	// 	if name.Valid {
	// 		util.Logger.Debug().Msgf("Section: %s", name.String)

	// 		// Marker can be in a paragraph, or in a heading. Either way, the containing node is the section starting node.
	// 		section := ast.NewSectionPointer(name.String)

	// 		if heading, ok := node.Parent().(*goldast.Heading); ok {
	// 			util.Logger.Debug().Msg("Parent is a heading")
	// 			heading.Parent().InsertBefore(heading.Parent(), heading, section)
	// 			section.StartNode = heading

	// 			return &SectionProcessor{SectionPointer: section}
	// 		} else {
	// 			node.Parent().InsertBefore(node.Parent(), node, section)
	// 			section.StartNode = node.NextSibling()

	// 			return &SectionProcessor{SectionPointer: section}
	// 		}

	// 	} else {
	// 		panic("Section doesn't have a name")
	// 	}

	// }

	if node.HasAttr("save") {
		if fcb, ok := nextNode.(*goldast.FencedCodeBlock); ok {
			if saveName := node.GetAttr("save"); saveName.Valid {
				executionBlock := ast.NewSaveCodeBlock(fcb, saveName.String)

				treatments.Ignore(node)
				treatments.Replace(nodeToReplace, executionBlock)
			}
		}

		return nil
	}

	if node.HasAttr("opt") {
		opt := ast.NewSectionOption(node.GetAttr("opt").String)
		opt.OptionRequired = node.HasAttr("required")
		opt.OptionPrompt = node.GetAttr("prompt")

		if node.HasAttr("type") {
			opt.OptionType = ast.BuildOptionType(node.GetAttr("type").String)
		} else {
			opt.OptionType = ast.BuildOptionType("string")
		}

		if node.HasAttr("default") {
			defaultVal := node.GetAttr("default")
			if defaultVal.Valid {
				if opt.OptionType.Validate(defaultVal.String) == nil {
					def := opt.OptionType.Normalise(defaultVal.String)
					opt.OptionDefault = null.StringFrom(def)
				} else {
					// ERROR: Default value isn't valid.
				}
			}
		}

		treatments.Replace(nodeToReplace, opt)

		return nil
	}

	if node.HasAttr("desc") {
		descNode := ast.NewDescriptionBlock()

		if node.ChildCount() > 0 {
			treatments.ReplaceWithChildren(node, descNode, node)
		} else if msg := node.GetAttr("desc"); msg.String != "" {
			fmt.Printf("Desc: %s\n", msg.String)
			node.AppendChild(node, goldast.NewString([]byte(msg.String)))
			treatments.ReplaceWithChildren(nodeToReplace, descNode, node)
		} else {
			panic("Bad desc node")
		}

		return nil
	}

	if node.HasAttr("stop-fail") {
		stop := ast.NewStopFail()

		if msg := node.GetAttr("stop-fail"); msg.Valid {
			node.AppendChild(node, goldast.NewString([]byte(msg.String)))
		}

		treatments.ReplaceWithChildren(nodeToReplace, stop, node)
		return nil
	}

	if node.HasAttr("stop-ok") {
		stop := ast.NewStopOk()

		if msg := node.GetAttr("stop-ok"); msg.Valid {
			node.AppendChild(node, goldast.NewString([]byte(msg.String)))
		}

		treatments.ReplaceWithChildren(nodeToReplace, stop, node)
		return nil
	}

	if node.HasAttr("ignore") {
		treatments.Remove(nodeToReplace)
		return nil
	}

	if node.HasAttr("on-failure") {
		fail := ast.NewOnFailure()
		fail.FailureMessageRegexp = node.GetAttr("on-failure").String

		treatments.ReplaceWithChildren(nodeToReplace, fail, node)

		return nil
	}

	if fcb, ok := nextNode.(*goldast.FencedCodeBlock); ok && node.HasAttr("with", "spinner", "stdout", "subenv", "sub-env", "capture-env", "replace", "borg", "reveal", "reveal-only") {
		executionBlock := ast.NewExecutionBlock(fcb)

		executionBlock.ShowStdout = node.HasAttr("stdout")
		executionBlock.ShowStderr = node.HasAttr("stderr")
		executionBlock.Reveal = node.HasAttr("reveal", "reveal-only")
		executionBlock.Execute = !node.HasAttr("reveal-only", "norun")
		executionBlock.SubstituteEnvironment = node.HasAttr("subenv") || node.HasAttr("sub-env")
		executionBlock.CaptureEnvironment = node.HasAttr("capture-env")
		executionBlock.ReplaceProcess = node.HasAttr("borg")

		if spinnerName := node.GetAttr("spinner"); spinnerName.Valid {
			executionBlock.SpinnerName = spinnerName.String
			executionBlock.SpinnerMode = ast.SpinnerModeVisible
		} else if node.HasAttr("nospin") {
			executionBlock.SpinnerMode = ast.SpinnerModeHidden
		} else if node.HasAttr("named") {
			executionBlock.SpinnerMode = ast.SpinnerModeInlineFirst
		} else if node.HasAttr("named-all") {
			executionBlock.SpinnerMode = ast.SpinnerModeInlineAll
		}

		if withVal := node.GetAttr("with"); withVal.Valid {
			executionBlock.With = withVal.String
		} else {
			executionBlock.With = string(fcb.Info.Text(reader.Source()))
		}

		// Execution block goes after the fenced code block, in case we're displaying the source.
		fcb.Parent().InsertAfter(fcb.Parent(), fcb, executionBlock)
		treatments.Remove(nodeToReplace)

		fmt.Printf("Created execution block.\n")
		executionBlock.Dump(reader.Source(), 5)
		fmt.Printf("Replacing:\n")
		nodeToReplace.Dump(reader.Source(), 5)

		fmt.Printf("Doc is:\n")
		node.OwnerDocument().Dump(reader.Source(), 5)

		if !executionBlock.Reveal {
			fmt.Printf("Removing the fenced code block, as we're not displaying it.\n")
			treatments.Remove(fcb)
		} else {
			treatments.Ignore(fcb)
		}

		return nil
	}

	if node.HasAttr("subenv", "sub-env") {
		fmt.Printf("HAVE subenv\n")
		t := NewTreatment(reader)
		goldast.Walk(node, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
			fmt.Printf("Walking %s\n", n.Kind().String())
			if entering {
				ConvertTextForSubenv(n, reader, t)
			}

			return goldast.WalkContinue, nil
		})
		t.Process(reader)

		treatments.DissolveRundownBlock(node)
	}

	return nil
}