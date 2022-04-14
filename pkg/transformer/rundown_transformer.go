package transformer

import (
	"fmt"
	"strings"

	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/util"
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
	Errors []error
}

// Rundown AST Transformer converts Rundown Elements in the markdown tree
// into proper rundown nodes, and applies any effects.
func NewRundownASTTransformer() *rundownASTTransformer {
	return &rundownASTTransformer{Errors: []error{}}
}

type OpenTags struct {
	data *RundownHtmlTag
	node goldast.Node
}

func createRundownBlocks(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var openNodes = []OpenTags{}

	// doc.Dump(reader.Source(), 0)

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

	// util.Logger.Trace().Msgf("Rundown Blocks:\n")
	// doc.Dump(reader.Source(), 0)
}

func (a *rundownASTTransformer) Transform(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	createRundownBlocks(doc, reader, pc)
	mergeTextBlocks(doc, reader, pc)
	a.convertRundownBlocks(doc, reader, pc)
}

// Merges sequential text nodes into a single text block. This makes subsequent processing easier.
func mergeTextBlocks(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		if text, ok := node.(*goldast.Text); ok {
			for {
				if nextText, ok := node.NextSibling().(*goldast.Text); ok && nextText.Segment.Start == text.Segment.Stop {
					text.Segment.Stop = nextText.Segment.Stop
					node.Parent().RemoveChild(node.Parent(), nextText)
				} else {
					break
				}
			}
		}

		return goldast.WalkContinue, nil
	})

}

func (a *rundownASTTransformer) convertRundownBlocks(doc *goldast.Document, reader goldtext.Reader, pc parser.Context) {
	var treatments *Treatment = NewTreatment(reader)
	var activeProcessors = []NodeProcessor{}

	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch node := node.(type) {

		case *ast.RundownBlock:
			processor, err := ConvertToRundownNode(node, reader, treatments)

			if err != nil {
				a.Errors = append(a.Errors, err)
			}

			if processor != nil {
				processor.Begin(node)
				activeProcessors = append(activeProcessors, processor)
			}

			// Don't leave the Rundown Element in the document tree.
			// treatments.Remove(node)

		case *goldast.FencedCodeBlock:
			// if !treatments.IsIgnored(node) {
			// 	eb := ast.NewExecutionBlock(node)
			// 	if with := node.Info; with != nil {
			// 		eb.With = string(with.Text(reader.Source()))
			// 	}
			// 	treatments.Replace(node, eb)
			// }

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

	// util.Logger.Trace().Msgf("Unprocessed doc is:\n")
	// doc.Dump(reader.Source(), 2)
	treatments.Process(reader)
	// util.Logger.Trace().Msgf("Processed doc is:\n")
	// doc.Dump(reader.Source(), 2)

	// Populate sections
	goldast.Walk(doc, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			if end, ok := n.(*ast.SectionEnd); ok {
				util.Logger.Trace().Msgf("Found section end\n")
				section := end.SectionPointer
				PopulateSectionMetadata(section, end, reader)
			}
		}

		return goldast.WalkContinue, nil
	})

	util.Logger.Trace().Msgf("Sections populated\n")
}

func ConvertToRundownNode(node *ast.RundownBlock, reader goldtext.Reader, treatments *Treatment) (NodeProcessor, error) {
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
			// para.Dump(reader.Source(), 0)
		}
	}

	if node.HasAttr("import") {
		prefix := node.GetAttr("import")
		importBlock := ast.NewImportBlock()

		if prefix.Valid {
			importBlock.ImportPrefix = prefix.String
		}

		treatments.ReplaceWithChildren(node, importBlock, node)
		return nil, nil
	}

	if node.HasAttr("label", "section") {
		name := node.GetAttr("section")
		if !name.Valid {
			name = node.GetAttr("label")
		}

		start := ast.NewSectionPointer(name.String)
		start.Silent = node.HasAttr("silent")

		if name.Valid {
			if heading, ok := parentNode.(*goldast.Heading); ok {
				start.StartNode = heading
				start.DescriptionShort = strings.Trim(string(heading.Text(reader.Source())), " ")

				heading.Parent().InsertBefore(heading.Parent(), heading, start)

				if ifScript := node.GetAttr("if"); ifScript.Valid {
					start.SetIfScript(ifScript.String)
				}

				util.Logger.Debug().Msgf("FindEndOfSection: %s", start.SectionName)
				end := FindEndOfSection(heading)
				if end != nil {
					end.Parent().InsertBefore(end.Parent(), end, start.End)
				} else {
					node.OwnerDocument().AppendChild(node.OwnerDocument(), start.End)
				}

				treatments.Remove(node)
			} else {
				// Section is not attached to a heading. Surround the Rundown Block with the section markers,
				// and dissolve the Rundown Block.
				start.StartNode = node.NextSibling()
				start.DescriptionShort = node.GetAttr("desc").ValueOrZero()

				node.Parent().InsertBefore(node.Parent(), node, start)
				node.Parent().InsertAfter(node.Parent(), node, start.End)

				treatments.DissolveRundownBlock(node)
			}
		}

		return nil, nil
	} else if h, ok := node.Parent().(*goldast.Heading); ok {
		// Check if there's a conditional...

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			cond := ast.NewConditionalStart()
			cond.SetIfScript(ifScript.String)

			h.Parent().InsertBefore(h.Parent(), h, cond)

			postH := false
			nextHeading := ast.FindNode(h.Parent(), func(n goldast.Node) bool {
				if n == h {
					postH = true
					return false
				}

				if postH {
					if hh, ok := n.(*goldast.Heading); ok {
						// The end is another heading at the same level, or a heading at a lower level.
						return hh.Level <= h.Level
					}
				}

				return false
			})

			// Make sure the conditional end stays inside the section.
			if nextHeading != nil {
				if n, ok := nextHeading.PreviousSibling().(*ast.SectionEnd); ok {
					nextHeading = n
				}

				nextHeading.Parent().InsertBefore(nextHeading.Parent(), nextHeading, cond.End)
			} else {
				h.OwnerDocument().AppendChild(h.OwnerDocument(), cond.End)
			}

			treatments.Remove(node)
		}
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

	if node.HasAttr("save") || node.HasAttr("save-as") {
		if fcb, ok := nextNode.(*goldast.FencedCodeBlock); ok {

			if saveName := node.GetFirstAttr("save", "save-as"); saveName.Valid {
				executionBlock := ast.NewSaveCodeBlock(fcb, saveName.String)
				executionBlock.Reveal = node.HasAttr("reveal")

				for _, r := range node.GetAttrList("replace") {
					rParts := strings.SplitN(r, ":", 2)
					if len(rParts) == 1 {
						executionBlock.Replacements[r] = r
					} else {
						executionBlock.Replacements[rParts[0]] = rParts[1]
					}
				}

				treatments.Ignore(node)
				treatments.Replace(nodeToReplace, executionBlock)
				treatments.Remove(fcb) // SaveCodeBlock looks after it's own rendering.
			}
		}

		return nil, nil
	}

	if node.HasAttr("opt") {
		opt := ast.NewSectionOption(node.GetAttr("opt").String)
		opt.OptionRequired = node.HasAttr("required")
		opt.OptionPrompt = node.GetAttr("prompt")
		opt.OptionDescription = node.GetAttr("desc").String
		opt.OptionTypeString = node.GetAttr("type").String

		if as := node.GetAttr("as"); as.Valid {
			opt.OptionAs = strings.ToUpper(as.String)
		}

		if node.HasAttr("type") {
			opt.OptionType = ast.BuildOptionType(node.GetAttr("type").String)
		} else {
			opt.OptionType = ast.BuildOptionType("string")
		}

		if opt.OptionType == nil {
			return nil, fmt.Errorf("unknown option type %s for option %s", opt.OptionTypeString, opt.OptionName)
		}

		if node.HasAttr("default") {
			defaultVal := node.GetAttr("default")
			if defaultVal.Valid {
				if opt.OptionType.Validate(defaultVal.String) == nil {
					def := opt.OptionType.Normalise(defaultVal.String)
					opt.OptionDefault = null.StringFrom(def)
				} else {
					return nil, fmt.Errorf("default option type \"%s\" is invalid for option \"%s\"", defaultVal.String, opt.OptionName)
				}
			}
		}

		treatments.Replace(nodeToReplace, opt)

		return nil, nil
	}

	if node.HasAttr("desc") {
		descNode := ast.NewDescriptionBlock()

		if node.ChildCount() > 0 {
			treatments.ReplaceWithChildren(node, descNode, node)
		} else if msg := node.GetAttr("desc"); msg.String != "" {
			util.Logger.Trace().Msgf("Desc: %s\n", msg.String)
			node.AppendChild(node, goldast.NewString([]byte(msg.String)))
			treatments.ReplaceWithChildren(nodeToReplace, descNode, node)
		} else {
			panic("Bad desc node")
		}

		return nil, nil
	}

	if node.HasAttr("help") {
		helpNode := ast.NewDescriptionBlock()

		if node.ChildCount() > 0 {
			treatments.ReplaceWithChildren(node, helpNode, node)
		}

		return nil, nil
	}

	if node.HasAttr("dep", "invoke") {
		invoke := ast.NewInvokeBlock()

		if dep := node.GetAttr("dep"); dep.Valid {
			invoke.Invoke = dep.String
			invoke.AsDependency = true
		} else if name := node.GetAttr("invoke"); name.Valid {
			invoke.Invoke = name.String
		}

		for _, attr := range node.Attrs {
			invoke.Args[attr.Key] = attr.Val
		}

		treatments.Replace(nodeToReplace, invoke)

		return nil, nil
	}

	if node.HasAttr("stop-fail") {
		stop := ast.NewStopFail()

		if msg := node.GetAttr("stop-fail"); msg.Valid {
			para := goldast.NewParagraph()
			para.AppendChild(para, goldast.NewString([]byte(msg.String)))
			node.AppendChild(node, para)
		}

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			stop.SetIfScript(ifScript.String)
		}

		treatments.ReplaceWithChildren(nodeToReplace, stop, node)
		return nil, nil
	}

	if node.HasAttr("stop-ok") {
		stop := ast.NewStopOk()

		if msg := node.GetAttr("stop-ok"); msg.Valid {
			para := goldast.NewParagraph()
			para.AppendChild(para, goldast.NewString([]byte(msg.String)))
			node.AppendChild(node, para)
		}

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			stop.SetIfScript(ifScript.String)
		}

		treatments.ReplaceWithChildren(nodeToReplace, stop, node)
		return nil, nil
	}

	if node.HasAttr("ignore") {
		treatments.Remove(nodeToReplace)
		return nil, nil
	}

	if node.HasAttr("on-failure") {
		fail := ast.NewOnFailure()
		fail.FailureMessageRegexp = node.GetAttr("on-failure").String

		treatments.ReplaceWithChildren(nodeToReplace, fail, node)

		return nil, nil
	}

	if fcb, ok := nextNode.(*goldast.FencedCodeBlock); ok && node.HasAttr("if", "with", "spinner", "stdout", "subenv", "sub-env", "capture-env", "replace", "borg", "reveal", "reveal-only", "skip-on-success") {
		executionBlock := ast.NewExecutionBlock(fcb)

		executionBlock.CaptureStdoutInto = node.GetAttr("stdout-into").String
		executionBlock.ShowStdout = node.HasAttr("stdout")
		executionBlock.ShowStderr = node.HasAttr("stderr")
		executionBlock.Reveal = node.HasAttr("reveal", "reveal-only")
		executionBlock.Execute = !node.HasAttr("reveal-only", "norun")
		executionBlock.SubstituteEnvironment = node.HasAttr("subenv") || node.HasAttr("sub-env")
		executionBlock.ReplaceProcess = node.HasAttr("borg")
		executionBlock.SkipOnSuccess = node.HasAttr("skip-on-success")
		executionBlock.SkipOnFailure = node.HasAttr("skip-on-failure")
		executionBlock.Language = string(fcb.Info.Text(reader.Source()))

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			executionBlock.SetIfScript(ifScript.String)
		}

		if envCapture := node.GetAttr("capture-env"); envCapture.Valid {
			executionBlock.CaptureEnvironment = strings.Split(envCapture.String, ",")

			for i := range executionBlock.CaptureEnvironment {
				executionBlock.CaptureEnvironment[i] = strings.TrimSpace(executionBlock.CaptureEnvironment[i])
			}
		}

		if spinnerName := node.GetAttr("spinner"); spinnerName.Valid {
			executionBlock.SpinnerName = spinnerName.String
			executionBlock.SpinnerMode = ast.SpinnerModeVisible
		}

		if node.HasAttr("nospin") {
			executionBlock.SpinnerMode = ast.SpinnerModeHidden
		} else if node.HasAttr("named") {
			executionBlock.SpinnerMode = ast.SpinnerModeInlineFirst
		} else if node.HasAttr("sub-spinners") || node.HasAttr("named-all") {
			executionBlock.SpinnerMode = ast.SpinnerModeInlineAll
		}

		if withVal := node.GetAttr("with"); withVal.Valid {
			executionBlock.With = withVal.String
		} else {
			executionBlock.With = executionBlock.Language
		}

		// Execution block goes after the fenced code block, in case we're displaying the source.
		if executionBlock.Execute {
			fcb.Parent().InsertAfter(fcb.Parent(), fcb, executionBlock)
		}
		treatments.Remove(nodeToReplace)

		// util.Logger.Trace().Msgf("Created execution block.\n")
		// executionBlock.Dump(reader.Source(), 5)
		// util.Logger.Trace().Msgf("Replacing:\n")
		// nodeToReplace.Dump(reader.Source(), 5)

		// util.Logger.Trace().Msgf("Doc is:\n")
		// node.OwnerDocument().Dump(reader.Source(), 5)

		if !executionBlock.Reveal {
			util.Logger.Trace().Msgf("Removing the fenced code block, as we're not displaying it.\n")
			treatments.Remove(fcb)
		} else {
			if node.HasAttr("sub-env") {
				wrapper := ast.NewSubEnvBlock(fcb)
				fcb.Parent().InsertAfter(fcb.Parent(), fcb, wrapper)
				wrapper.AppendChild(wrapper, fcb)
			}

			treatments.Ignore(fcb)

		}

		return nil, nil
	}

	if node.HasAttr("subenv", "sub-env") {

		util.Logger.Trace().Msgf("HAVE subenv\n")
		t := NewTreatment(reader)
		goldast.Walk(node, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
			util.Logger.Trace().Msgf("Walking %s\n", n.Kind().String())
			if entering {
				ConvertTextForSubenv(n, reader, t)
			}

			return goldast.WalkContinue, nil
		})
		t.Process(reader)

		treatments.DissolveRundownBlock(node)
	}

	return nil, nil
}
