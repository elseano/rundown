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
	var openNodes = []OpenTags{}

	processed := []goldast.Node{}

	// doc.Dump(reader.Source(), 0)

	// First, transform rundown opening/closing RawHTML into RundownBlocks.
	// This makes the next phase simpler in terms of handling what's inside any rundown block forms.
	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch node := node.(type) {

		case *goldast.RawHTML, *goldast.HTMLBlock:
			for _, htmlNode := range ExtractRundownElement(node, reader, "") {
				fmt.Printf("Processing %+v\n", htmlNode)

				if htmlNode.closed {
					// No content.
					rdb := ast.NewRundownBlock()
					rdb.Attrs = htmlNode.attrs
					rdb.TagName = htmlNode.tag

					node.Parent().InsertBefore(node.Parent(), node, rdb)
					processed = append(processed, node)
				} else if htmlNode.closer {
					// Create block.
					if len(openNodes) > 0 {
						openingElement := openNodes[len(openNodes)-1]
						openNodes = openNodes[0 : len(openNodes)-1]

						rdb := ast.NewRundownBlock()
						rdb.Attrs = openingElement.data.attrs
						rdb.TagName = openingElement.data.tag

						fmt.Printf("Found a closing node of a block\nMoving children into rdb")

						// Move all nodes between start and end into rdb.
						var nextChild goldast.Node
						for child := openingElement.node.NextSibling(); child != nil && child != node; child = nextChild {
							fmt.Printf("Adding %s\n", child.Kind().String())
							nextChild = child.NextSibling()
							rdb.AppendChild(rdb, child)
						}

						fmt.Printf("Inserting before %s\n", openingElement.node.Kind().String())
						openingElement.node.Parent().InsertBefore(openingElement.node.Parent(), openingElement.node, rdb)
						processed = append(processed, openingElement.node)
						processed = append(processed, node)
					}
				} else {
					openNodes = append(openNodes, OpenTags{data: htmlNode, node: node})
				}

			}
		}

		return goldast.WalkContinue, nil
	})

	for _, n := range processed {
		if n.Parent() != nil {
			n.Parent().RemoveChild(n.Parent(), n)
		}
	}

	// util.Logger.Trace().Msgf("Rundown Blocks:\n")
	doc.Dump(reader.Source(), 0)
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

func (a *rundownASTTransformer) convertRundownBlocks(doc goldast.Node, reader goldtext.Reader, pc parser.Context) {
	util.Logger.Debug().Msg(util.CaptureStdout(func() {
		doc.Dump(reader.Source(), 0)
	}))

	// Because the AST will change significantly, we can't use the goldmark Walk function.
	for node := doc.FirstChild(); node != nil; {
		util.Logger.Debug().Msgf("Converting loop on %T", node)

		switch n := node.(type) {

		case *ast.RundownBlock:
			var err error
			node, err = ConvertToRundownNode(n, reader)

			if err != nil {
				a.Errors = append(a.Errors, err)
			}

			util.Logger.Debug().Msgf("AST is now: \n%s", util.CaptureStdout(func() {
				doc.Dump(reader.Source(), 0)
			}))

			util.Logger.Debug().Msgf("Returned node is %T", node)
		}

		if node != nil {
			if node.ChildCount() > 0 {
				// If there are children on this node, lets dive into them.
				node = node.FirstChild()
			} else if node.NextSibling() != nil {
				// Otherwise, if theres a next sibling, examine that.
				node = node.NextSibling()
			} else if node.Parent() != nil {
				// If no next sibling, walk up the parents until there's a next sibling there.
				for node != nil && node.NextSibling() == nil {
					node = node.Parent()
				}

				if node != nil {
					node = node.NextSibling()
				}
			} else {
				node = nil
			}
		}

		util.Logger.Debug().Msgf("Next node is now: %T", node)

	}

	// Populate sections
	goldast.Walk(doc, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			if section, ok := n.(*ast.SectionPointer); ok {
				util.Logger.Trace().Msgf("Found section end\n")
				ast.FillInvokeBlocks(section, 10)
				PopulateSectionMetadata(section, reader)
			}
		}

		return goldast.WalkContinue, nil
	})

	util.Logger.Trace().Msgf("Sections populated\n")
}

// Converts a RundownBlock into a proper instruction node. Returns the node to continue iterating from, or an error.
func ConvertToRundownNode(node *ast.RundownBlock, reader goldtext.Reader) (goldast.Node, error) {
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

		ReplaceWithChildren(node, importBlock, node)
		return importBlock, nil
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
				util.Logger.Debug().Msgf("New Section: %s", start.SectionName)

				start.StartNode = heading
				start.DescriptionShort = strings.Trim(string(heading.Text(reader.Source())), " ")

				heading.Parent().InsertBefore(heading.Parent(), heading, start)

				if ifScript := node.GetAttr("if"); ifScript.Valid {
					start.SetIfScript(ifScript.String)
				}

				end := FindEndOfSection(heading)
				util.Logger.Debug().Msgf("FindEndOfSection %s is %T", start.SectionName, end)

				for child := start.StartNode; child != nil; {
					nextChild := child.NextSibling()

					util.Logger.Debug().Msgf("Adding %T to section %s", child, start.SectionName)

					AppendChild(start, child)

					if child == end {
						break
					}

					child = nextChild
				}

				Remove(node, reader)

				return start, nil
			}
		}

		return node, nil
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

			// // Make sure the conditional end stays inside the section.
			if nextHeading != nil {
				// 	if n, ok := nextHeading.PreviousSibling().(*ast.SectionEnd); ok {
				// 		nextHeading = n
				// 	}

				nextHeading.Parent().InsertBefore(nextHeading.Parent(), nextHeading, cond.End)
			} else {
				h.OwnerDocument().AppendChild(h.OwnerDocument(), cond.End)
			}
			Remove(node, reader)
			return cond, nil
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

				Replace(nodeToReplace, executionBlock)
				Remove(fcb, reader)
				return executionBlock, nil // SaveCodeBlock looks after it's own rendering.
			}
		}

		return node, nil
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
			return node, fmt.Errorf("unknown option type %s for option %s", opt.OptionTypeString, opt.OptionName)
		}

		if node.HasAttr("default") {
			defaultVal := node.GetAttr("default")
			if defaultVal.Valid {
				if opt.OptionType.Validate(defaultVal.String) == nil {
					def := opt.OptionType.Normalise(defaultVal.String)
					opt.OptionDefault = null.StringFrom(def)
				} else {
					return node, fmt.Errorf("default option type \"%s\" is invalid for option \"%s\"", defaultVal.String, opt.OptionName)
				}
			}
		}

		Replace(nodeToReplace, opt)

		return opt, nil
	}

	if node.HasAttr("desc") {
		descNode := ast.NewDescriptionBlock()

		if node.ChildCount() > 0 {
			ReplaceWithChildren(node, descNode, node)
		} else if msg := node.GetAttr("desc"); msg.String != "" {
			util.Logger.Trace().Msgf("Desc: %s\n", msg.String)
			node.AppendChild(node, goldast.NewString([]byte(msg.String)))
			ReplaceWithChildren(nodeToReplace, descNode, node)
		} else {
			panic("Bad desc node")
		}

		return descNode, nil
	}

	if node.HasAttr("help") {
		helpNode := ast.NewDescriptionBlock()

		ReplaceWithChildren(node, helpNode, node)

		return helpNode, nil
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

		Replace(nodeToReplace, invoke)

		return invoke, nil
	}

	if node.HasAttr("stop-fail") {
		stop := ast.NewStopFail()

		if msg := node.GetAttr("stop-fail"); msg.Valid && msg.String != "" {
			para := goldast.NewParagraph()
			para.AppendChild(para, goldast.NewString([]byte(msg.String)))
			node.AppendChild(node, para)
		}

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			stop.SetIfScript(ifScript.String)
		}

		ReplaceWithChildren(nodeToReplace, stop, node)
		return stop, nil
	}

	if node.HasAttr("stop-ok") {
		stop := ast.NewStopOk()

		if msg := node.GetAttr("stop-ok"); msg.Valid && msg.String != "" {
			para := goldast.NewParagraph()
			para.AppendChild(para, goldast.NewString([]byte(msg.String)))
			node.AppendChild(node, para)
		}

		if ifScript := node.GetAttr("if"); ifScript.Valid {
			stop.SetIfScript(ifScript.String)
		}

		ReplaceWithChildren(nodeToReplace, stop, node)
		return stop, nil
	}

	if node.HasAttr("ignore") {
		return Remove(nodeToReplace, reader), nil
	}

	if node.HasAttr("on-failure") {
		fail := ast.NewOnFailure()
		fail.FailureMessageRegexp = node.GetAttr("on-failure").String

		ReplaceWithChildren(nodeToReplace, fail, node)

		return fail, nil
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
		Remove(nodeToReplace, reader)

		// util.Logger.Trace().Msgf("Created execution block.\n")
		// executionBlock.Dump(reader.Source(), 5)
		// util.Logger.Trace().Msgf("Replacing:\n")
		// nodeToReplace.Dump(reader.Source(), 5)

		// util.Logger.Trace().Msgf("Doc is:\n")
		// node.OwnerDocument().Dump(reader.Source(), 5)

		if !executionBlock.Reveal {
			util.Logger.Trace().Msgf("Removing the fenced code block, as we're not displaying it.\n")
			Remove(fcb, reader)
		} else {
			if node.HasAttr("sub-env") {
				wrapper := ast.NewSubEnvBlock(fcb)
				fcb.Parent().InsertAfter(fcb.Parent(), fcb, wrapper)
				wrapper.AppendChild(wrapper, fcb)
			}

		}

		return executionBlock, nil
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

		from := node.PreviousSibling()
		DissolveRundownBlock(node)
		return from, nil
	}

	return node, nil
}
