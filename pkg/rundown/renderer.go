package rundown

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/elseano/rundown/pkg/markdown"
	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func PrepareMarkdown() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(
			markdown.ConsoleRenderer,
			extension.GFM,
			markdown.CodeModifiers,
			markdown.InvisibleBlocks,
			extension.Strikethrough,
			markdown.RundownElements,
			markdown.Emoji,
			// CodeExecute,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return md
}

var hasDarkBkg = false

func init() {
	hasDarkBkg = termenv.HasDarkBackground()
}

func NewRenderer() (goldmark.Markdown, *RundownRenderer) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Strikethrough,
			markdown.InvisibleBlocks,
			markdown.RundownElements,
			markdown.Emoji,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	ansiOptions := ansi.Options{
		WordWrap:     80,
		ColorProfile: termenv.TrueColor,
	}

	if hasDarkBkg {
		ansiOptions.Styles = glamour.DarkStyleConfig
	} else {
		ansiOptions.Styles = glamour.LightStyleConfig
	}

	ansiOptions.ColorProfile = termenv.ColorProfile()

	ar := ansi.NewRenderer(ansiOptions)
	rd := &RundownRenderer{ctx: NewContext(), ar: ar}
	rd.ctx.RawOut = os.Stdout

	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(
				util.Prioritized(ar, 1000),
				util.Prioritized(rd, 1),
			),
		),
	)
	return md, rd
}

type RundownRenderer struct {
	ctx         *Context
	ar          *ansi.ANSIRenderer
	docRenderer renderer.NodeRendererFunc
}

type DummyRegister struct {
	docRenderer renderer.NodeRendererFunc
}

func (d *DummyRegister) Register(k ast.NodeKind, fun renderer.NodeRendererFunc) {
	if k == ast.KindDocument {
		d.docRenderer = fun
	}
}

func (r *RundownRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Hacky, but need to capture document renderer.
	d := &DummyRegister{}
	r.ar.RegisterFuncs(d)
	r.docRenderer = d.docRenderer

	// blocks
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(markdown.KindRundownBlock, r.renderRundownBlock)
	reg.Register(markdown.KindExecutionBlock, r.renderExecutionBlock)
	reg.Register(markdown.KindSection, r.renderSection)
	reg.Register(markdown.KindSectionedDocument, r.renderSectionedDocument)

	// inline
	reg.Register(markdown.KindEmojiInline, r.renderEmoji)
	reg.Register(markdown.KindRundownInline, r.renderRundownInline)
}

func deleteForward(node ast.Node) {
	if node == nil {
		return
	}

	deleteForward(node.Parent())

	for n := node.NextSibling(); n != nil; {
		n2 := n.NextSibling()
		n.Parent().RemoveChild(n.Parent(), n)
		n = n2
	}
}

func (r *RundownRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Don't render the outer document.
	if _, ok := node.Parent().(*markdown.SectionedDocument); ok {
		return ast.WalkContinue, nil
	} else {
		return r.docRenderer(w, source, node, entering)
	}
}

func (r *RundownRenderer) renderRundownBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	if entering {
		if rundown, ok := node.(*markdown.RundownBlock); ok {
			// We don't render function/shortcode option contents, thats for reading only.
			if rundown.Modifiers.HasAny("ignore", "opt") {
				return ast.WalkSkipChildren, nil
			}

			if rundown.GetModifiers().Flags[StopOkFlag] {
				deleteForward(node)

				return ast.WalkContinue, nil
			}

			if rundown.GetModifiers().Flags[StopFailFlag] {
				deleteForward(node)

				// return ast.WalkStop, &StopError{Result: StopFailResult}
				return ast.WalkContinue, nil
			}

			if message, ok := rundown.GetModifiers().Values[StopFailParameter]; ok {
				deleteForward(node)

				result := StopFailResult
				result.Message = message

				r.ctx.SetError(&StopError{Result: result})

				return ast.WalkContinue, nil
			}

			if setEnv, ok := rundown.GetModifiers().Flags[markdown.Flag("set-env")]; setEnv && ok {
				for k, val := range rundown.GetModifiers().Values {
					envName := strings.ReplaceAll(strings.ToUpper(string(k)), "-", "_")
					r.ctx.SetEnv(envName, val)
				}
			}

			if rundown.Modifiers.HasAny("on-failure") {
				if val, ok := rundown.Modifiers.Flags[markdown.Flag("on-failure")]; val && ok && r.ctx.CurrentError != nil {
					return ast.WalkContinue, nil
				}

				if val, ok := rundown.Modifiers.Values[markdown.Parameter("on-failure")]; ok && r.ctx.CurrentError != nil {
					if stopError, ok := r.ctx.CurrentError.(*StopError); ok {
						r, _ := regexp.Compile(val)
						if r.MatchString(stopError.Result.Output) {
							return ast.WalkContinue, nil
						}
					}
				}

				return ast.WalkSkipChildren, nil
			}

			if rundown.GetModifiers().HasAny("invoke") {
				source := rundown.GetModifiers().GetValue("from")

				if source == nil {
					source = &r.ctx.CurrentFile
				}

				name := rundown.GetModifiers().GetValue("invoke")
				if name == nil {
					return ast.WalkStop, errors.New("Invoke requires a ShortCode value")
				}

				rd, err := LoadFile(*source)
				if err != nil {
					return ast.WalkStop, err
				}

				if info := rd.GetShortCodes().Functions[*name]; info != nil {
					section := info.Section

					mods := markdown.NewModifiers()
					mods.Flags[markdown.Flag("set-env")] = true

					for k, v := range rundown.GetModifiers().Values {
						if strings.HasPrefix(string(k), "opt-") {
							mods.Values[k] = v
						}
					}

					// Create a rundown block which sets the environment to the invoke options.
					envNode := markdown.NewRundownBlock(mods)

					node.Parent().InsertAfter(node.Parent(), node, section)

					// Adjust the section contents to be relative to the current level.
					parentSection := node.Parent()
					for {
						if _, ok := parentSection.(*markdown.Section); ok {
							break
						}

						parentSection = parentSection.Parent()
					}

					section.ForceLevel(parentSection.(*markdown.Section).Level)

					// Remove the heading when invoking functions, unless we specify we want to keep the heading
					if keepHeading, specified := mods.Flags[markdown.Flag("keep-heading")]; keepHeading == false || !specified {
						section.RemoveChild(section, section.FirstChild())
					}

					// Add the environment setting. FIXME - We should nest the Section inside this node to wrap the context.
					section.InsertBefore(section, section.FirstChild(), envNode)
				} else {
					// ShortCode not found in file.
					if flag, ok := rundown.GetModifiers().Flags["ignore-missing"]; flag && ok {
						return ast.WalkSkipChildren, nil
					}

					return ast.WalkStop, errors.New("Cannot find " + *name + " in " + *source)
				}

			}
		}
	}

	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderSectionedDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderExecutionBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	if entering {
		return ast.WalkSkipChildren, nil
	}

	executionBlock := node.(*markdown.ExecutionBlock)

	// We write to RawOut here, as out will be the word wrap writer.
	result := Execute(r.ctx, executionBlock, source, r.ctx.Logger, r.ctx.RawOut)
	if _, ok := node.NextSibling().(*markdown.ExecutionBlock); !ok {
		w.WriteString("\n")
	}

	switch result {
	case SuccessfulExecution:
		return ast.WalkContinue, nil
	case SkipToNextHeading:

		nextNode := node.NextSibling()

		deleteNodesUntilSection(nextNode)

		return ast.WalkContinue, nil
	case StopOkResult:
		deleteForward(node)
		return ast.WalkContinue, nil
	case StopFailResult:
		deleteForward(node)
		return ast.WalkContinue, errors.New("Stop requested")
	}

	// Otherwise, error.

	deleteForward(node)

	var section *markdown.Section
	var n ast.Node

	for n = node; section == nil && n != nil; n = n.Parent() {
		if s, ok := n.Parent().(*markdown.Section); ok {
			section = s
		}
	}

	r.ctx.SetError(&StopError{Result: result, StopHandlers: section.Handlers})

	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderSection(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderEmoji(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderRundownInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	return ast.WalkContinue, nil
}
