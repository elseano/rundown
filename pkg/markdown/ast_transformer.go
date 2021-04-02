package markdown

import (
	"container/list"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

/*
 *
 * Rundown AST Transformer
 * - Moves FCB modifiers into dedicated Rundown blocks.
 * - Handles loose RundownBlocks
 * - Builds Section container nodes & rearranges handlers.
 * - Builds ExecutionBlock nodes.
 *
 */

type rundownASTTransformer struct {
}

var defaultRundownASTTransformer = &rundownASTTransformer{}

// NewFootnoteASTTransformer returns a new parser.ASTTransformer that
// insert a footnote list to the last of the document.
func NewRundownASTTransformer() parser.ASTTransformer {
	return defaultRundownASTTransformer
}

func getRawText(n ast.Node, source []byte) string {
	result := ""
	for i := 0; i < n.Lines().Len(); i++ {
		s := n.Lines().At(i)
		result = result + string(source[s.Start:s.Stop])
	}
	return result
}

var isRundownOpening = regexp.MustCompile(`<(?:r|rundown)\s+(.*?)\s*(?:/>|>)`)
var isRundownClosing = regexp.MustCompile(`</(?:r|rundown)>`)

func (a *rundownASTTransformer) fixConsecutiveTexts(doc *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		text, isText := node.(*ast.Text)

		if isText {
			for {
				if next, nextText := node.NextSibling().(*ast.Text); nextText {
					if text.Merge(next, reader.Source()) {
						node.Parent().RemoveChild(node.Parent(), next)
					} else {
						break
					}
				} else {
					break
				}
			}
		}

		return ast.WalkContinue, nil
	})
}

func (a *rundownASTTransformer) fixTrailingHeadingSpace(doc *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if rd, ok := node.(*RundownInline); ok && rd.Modifiers.HasAny("label", "func") {
			_, hOk := node.Parent().(*ast.Heading)
			t, tOk := node.PreviousSibling().(*ast.Text)

			if hOk && tOk {
				trimmed := t.Segment.TrimRightSpace(reader.Source())
				t.Segment = trimmed
			}
		}
		return ast.WalkContinue, nil
	})
}

// func (a *rundownASTTransformer) injectCallSites(doc *ast.Document, reader text.Reader, pc parser.Context) {
// 	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
// 		if rundown, ok := node.(*RundownInline); ok {

// 			if rundown.GetModifiers().HasAny("invoke") {
// 				source := rundown.GetModifiers().GetValue("from")

// 				if source == nil {
// 					source = &r.ctx.CurrentFile
// 				}

// 				name := rundown.GetModifiers().GetValue("invoke")
// 				if name == nil {
// 					return ast.WalkStop, errors.New("Invoke requires a ShortCode value")
// 				}

// 				rd, err := LoadFile(*source)
// 				if err != nil {
// 					return ast.WalkStop, err
// 				}

// 				if info := rd.GetShortCodes().Functions[*name]; info != nil {
// 					section := info.Section

// 					mods := markdown.NewModifiers()
// 					mods.Flags[markdown.Flag("set-env")] = true

// 					for k, v := range rundown.GetModifiers().Values {
// 						if strings.HasPrefix(string(k), "opt-") {
// 							mods.Values[k] = v
// 						}
// 					}

// 					// Create a rundown block which sets the environment to the invoke options.
// 					envNode := markdown.NewRundownBlock(mods)

// 					node.Parent().InsertAfter(node.Parent(), node, section)

// 					// Adjust the section contents to be relative to the current level.
// 					parentSection := node.Parent()
// 					for {
// 						if _, ok := parentSection.(*markdown.Section); ok {
// 							break
// 						}

// 						parentSection = parentSection.Parent()
// 					}

// 					section.ForceLevel(parentSection.(*markdown.Section).Level)

// 					// Remove the heading when invoking functions, unless we specify we want to keep the heading
// 					if keepHeading, specified := mods.Flags[markdown.Flag("keep-heading")]; keepHeading == false || !specified {
// 						section.RemoveChild(section, section.FirstChild())
// 					}

// 					// Add the environment setting. FIXME - We should nest the Section inside this node to wrap the context.
// 					section.InsertBefore(section, section.FirstChild(), envNode)
// 				} else {
// 					// ShortCode not found in file.
// 					if flag, ok := rundown.GetModifiers().Flags["ignore-missing"]; flag && ok {
// 						return ast.WalkSkipChildren, nil
// 					}

// 					return ast.WalkStop, errors.New("Cannot find " + *name + " in " + *source)
// 				}

// 			}
// 		}

// 		return ast.WalkContinue, nil
// 	})
// }

func (a *rundownASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	a.fixConsecutiveTexts(doc, reader, pc)
	a.fixTrailingHeadingSpace(doc, reader, pc)
	// Finds FencedCodeBlocks, and transforms their syntax line additions into RundownBlock elements
	// which provides consistency later.

	// Also finds HTMLBlocks which are rundown start and end tags, and everything inbetween into
	// a RundownBlock tag.

	// Also finds Paragraphs which have only one non-display RundownInline as a child, and converts to RundownBlock.

	// Also finds Headings which have labels/funcs and trims trailing space.

	var startBlock *ast.HTMLBlock
	var mods *Modifiers
	var nextNode ast.Node

	for node := doc.FirstChild(); node != nil; node = nextNode {
		nextNode = node.NextSibling()

		if html, ok := node.(*ast.HTMLBlock); ok && startBlock == nil {
			contents := getRawText(html, reader.Source())
			if match := isRundownOpening.FindStringSubmatch(contents); match != nil {
				mods = ParseModifiers(string(match[1]), "=")
				startBlock = html
			}
		} else if html, ok := node.(*ast.HTMLBlock); ok && startBlock != nil && isRundownClosing.MatchString(getRawText(html, reader.Source())) {
			rundown := NewRundownBlock(mods)
			doc.InsertBefore(doc, startBlock, rundown)

			for content := startBlock.NextSibling(); content != nil && content != html; {
				thisContent := content
				content = content.NextSibling()

				rundown.AppendChild(rundown, thisContent)
			}

			doc.RemoveChild(doc, startBlock)
			doc.RemoveChild(doc, html)

			startBlock = nil
			mods = nil
		} else if fcb, ok := node.(*ast.FencedCodeBlock); ok {
			// If this is a Fenced Code Block, and has an syntax specified, then
			// create an ExecutionBlock.

			var infoText string = ""

			info := node.(*ast.FencedCodeBlock).Info

			if info != nil {
				infoText = strings.TrimSpace(string(info.Text(reader.Source())))
				splitInfo := ""
				split := strings.SplitN(infoText, " ", 2)

				if len(split) == 2 { // Trim the syntax specifier
					fcb.Info.Segment = fcb.Info.Segment.WithStop(fcb.Info.Segment.Start + len(split[0]))
					splitInfo = split[1]
				}

				fencedMods := ParseModifiers(splitInfo, ":") // Fenced modifiers separate KV's with :

				if len(splitInfo) == 0 {
					// If there's no modifiers on the FCB, then check for a preceding Rundown block.
					if rdb, ok := node.PreviousSibling().(*RundownBlock); ok {

						// Make sure this rundown block is tagged with things for a FCB.
						if rdb.Modifiers.HasAny("stop-ok", "stop-fail", "desc", "opt") == false {
							fencedMods.Ingest(rdb.Modifiers)
							rdb.Parent().RemoveChild(rdb.Parent(), rdb)
						}
					}
				}

				// Insert the Execution Block after the FCB if we're running this thing.
				if len(split[0]) > 0 && fencedMods.Flags[Flag("norun")] != true {
					eb := NewExecutionBlock(split[0], fencedMods)
					eb.SetLines(fcb.Lines())
					eb.SetOrigin(fcb, reader.Source())
					fcb.Parent().InsertAfter(fcb.Parent(), fcb, eb)
				}

				// If the FCB isn't set to reveal, delete it.
				if fencedMods.Flags[Flag("reveal")] != true {
					fcb.Parent().RemoveChild(fcb.Parent(), fcb)
				}

			}
		} else if p, ok := node.(*ast.Paragraph); ok && p.ChildCount() == 1 {
			// Convert Paragraph > RundownInline into RundownBlock > Paragraph.
			// This makes the case of a Rundown Paragraph more obvious and easier to detect.
			if rundown, ok := p.FirstChild().(*RundownInline); ok {

				if !rundown.GetModifiers().HasAll("sub-env") {
					rundownBlock := NewRundownBlock(rundown.Modifiers)
					innerP := ast.NewParagraph()

					for c := rundown.FirstChild(); c != nil; {
						c2 := c
						c = c.NextSibling()
						innerP.AppendChild(rundownBlock, c2)
					}
					rundownBlock.AppendChild(rundownBlock, innerP)
					doc.ReplaceChild(doc, p, rundownBlock)
				}

			}
		}

	}

	a.TransformSections(doc, reader, pc)
}

func (a *rundownASTTransformer) TransformSections(doc *ast.Document, reader text.Reader, pc parser.Context) {
	toc := NewSectionedDocument()
	var currentSection *Section = nil
	var currentSectionTree = map[int]*Section{}
	var subjectNodes list.List

	// Because we're rearranging the AST, we're going to capture the list of block elements first.
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		subjectNodes.PushBack(node)
	}

	// Now traverse the elements, shunting them into Sections
	for nodeE := subjectNodes.Front(); nodeE != nil; nodeE = nodeE.Next() {
		node := nodeE.Value.(ast.Node)

		if heading, ok := node.(*ast.Heading); ok {
			section := NewSectionFromHeading(heading, reader.Source())
			toc.AddSection(section)

			var parent ast.Node = doc

			if p, ok := currentSectionTree[section.Level-1]; ok {
				parent = p
			}

			parent.AppendChild(parent, section)

			currentSection = section

			// Fill all sections downwards to be the currentSection.
			// This allows us to pick the correct parent when headings skip levels.
			currentSectionTree[section.Level] = section
			currentSectionTree[section.Level+1] = section
			currentSectionTree[section.Level+2] = section
			currentSectionTree[section.Level+3] = section
		} else if currentSection == nil {
			// Special case - we have content without a heading.
			// Create a root section, which allows us to have document-wide options.

			section := NewSectionForRoot()
			toc.AddSection(section)
			doc.AppendChild(doc, section)
			currentSection = section

			currentSectionTree[section.Level] = section
			currentSectionTree[section.Level+1] = section
			currentSectionTree[section.Level+2] = section
			currentSectionTree[section.Level+3] = section
			currentSectionTree[section.Level+4] = section
		}

		if currentSection != nil {
			currentSection.Append(node)
		}
	}

	if len(toc.Sections) > 0 {
		toc.AppendChild(toc, doc)
	}
}
