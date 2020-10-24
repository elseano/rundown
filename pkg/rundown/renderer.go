package rundown

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	icolor "image/color"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/elseano/rundown/pkg/markdown"
	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/kyokomi/emoji"
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

func strPtr(blah string) *string { return &blah }

func NewRenderer(ctx *Context) (goldmark.Markdown, *RundownRenderer) {
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
		WordWrap:     ctx.ConsoleWidth,
		ColorProfile: termenv.TrueColor,
	}

	if hasDarkBkg {
		ansiOptions.Styles = glamour.DarkStyleConfig
		ansiOptions.Styles.Document.Color = strPtr("255") // Increase contrast
	} else {
		ansiOptions.Styles = glamour.LightStyleConfig
		ansiOptions.Styles.Document.Color = strPtr("232") // Increase contrast
	}

	ansiOptions.ColorProfile = termenv.ColorProfile()

	ar := ansi.NewRenderer(ansiOptions)
	rd := &RundownRenderer{ctx: ctx, ar: ar}
	rd.ctx.Style = &ansiOptions.Styles
	rd.ctx.Profile = ansiOptions.ColorProfile
	rd.ctx.Renderer = renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(ar, 1000),
			util.Prioritized(rd, 1),
		),
	)

	ar.Register(markdown.KindRundownInline, NewRundownInlineBuilder(rd.ctx))
	ar.Register(ast.KindString, NewStringBuilder(rd.ctx))

	md.SetRenderer(rd.ctx.Renderer)

	return md, rd
}

func RenderToString(contents string) string {
	md, _ := NewRenderer(NewContext())
	var buf bytes.Buffer
	md.Convert([]byte(contents), &buf)
	return buf.String()
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

type StringBuilder struct {
	ansi.BaseElementBuilder
	ctx *Context
}

func NewStringBuilder(ctx *Context) *StringBuilder {
	return &StringBuilder{
		BaseElementBuilder: ansi.BaseElementBuilder{},
		ctx:                ctx,
	}
}

func (b *StringBuilder) NewElement(node ast.Node, source []byte, ctx *ansi.RenderContext) *ansi.Element {
	n := node.(*ast.String)
	s := string(n.Value)

	if n.Parent().Kind() == ast.KindEmphasis {
		return nil
	}

	return &ansi.Element{
		Renderer: &ansi.BaseElement{
			Token: html.UnescapeString(s),
			Style: ctx.Options.Styles.Text,
		},
	}

}

type RundownInlineBuilder struct {
	ansi.BaseElementBuilder
	ctx *Context
}

func NewRundownInlineBuilder(ctx *Context) *RundownInlineBuilder {
	return &RundownInlineBuilder{
		BaseElementBuilder: ansi.BaseElementBuilder{},
		ctx:                ctx,
	}
}

type SubElement struct {
	ansi.BlockElement
	ctx   *Context
	First bool
}

func (s SubElement) Finish(w io.Writer, ctx ansi.RenderContext) error {
	var buf bytes.Buffer
	s.BlockElement.Finish(&buf, ctx)
	// str, err := SubEnv(buf.String(), s.ctx)
	w.Write([]byte("Blah"))

	return nil
}

func (b *RundownInlineBuilder) NewElement(node ast.Node, source []byte, ctx *ansi.RenderContext) *ansi.Element {

	e := ansi.BlockElement{
		Block: &bytes.Buffer{},
		Style: ctx.BlockStack.Current().Style,
	}

	ee := SubElement{
		BlockElement: e,
		ctx:          b.ctx,
	}

	return &ansi.Element{
		Renderer: &ee.BlockElement,
		Finisher: ee,
	}
}

func (e *SubElement) Render(w io.Writer, ctx ansi.RenderContext) error {
	bs := ctx.BlockStack

	if !e.First {
		_, _ = w.Write([]byte("\n"))
	}
	be := ansi.BlockElement{
		Block: &bytes.Buffer{},
		Style: bs.Current().Style,
	}
	bs.Push(be)

	return nil
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
	reg.Register(markdown.KindEmojiInline, r.ar.RenderWrapper(r.renderEmoji))
	reg.Register(markdown.KindRundownInline, r.renderRundownInline)
	reg.Register(markdown.KindTextSubInline, r.ar.RenderNode)
	reg.Register(ast.KindImage, r.renderImage)
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

var urlMatch = regexp.MustCompile("^http(s?)://")

func (r *RundownRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if val, set := os.LookupEnv("COLORTERM"); !set || (val != "truecolor" && val != "24bit") {
		// Only render images when we're sure we're in truecolor mode.
		return r.ar.RenderNode(w, source, node, entering)
	}

	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)
	var marginInt = 0
	if r.ctx.Style.Document.Margin != nil {
		marginInt = int(*r.ctx.Style.Document.Margin)
	}

	var image *ansimage.ANSImage
	var maxWidth = r.ctx.ConsoleWidth - (marginInt * 2)

	if urlMatch.Match(n.Destination) {
		if i, err := ansimage.NewScaledFromURL(string(n.Destination), 40, maxWidth, icolor.Black, ansimage.ScaleModeFit, ansimage.NoDithering); err != nil {
			_, _ = w.WriteString("Error: " + err.Error())
		} else {
			image = i
		}
	} else {
		if i, err := ansimage.NewScaledFromFile(string(n.Destination), 40, maxWidth, icolor.Black, ansimage.ScaleModeFit, ansimage.NoDithering); err != nil {
			_, _ = w.WriteString("Error: " + err.Error())
		} else {
			image = i
		}
	}

	w.Flush() // Start with empty.

	if image != nil {
		marginLeft := strings.Repeat(" ", marginInt)
		lines := strings.Split(image.Render(), "\n")
		for _, line := range lines {
			if line != "" {
				w.WriteString(marginLeft + line + "\n")
			}

			// Must flush every line to avoid filling buffer.
			// Otherwise WordWrapWriter gets confused by half-formed escape codes on buffer full.
			w.Flush()
		}
	}

	w.WriteString("\n")

	return ast.WalkSkipChildren, nil

}
func (r *RundownRenderer) renderRundownBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Flush()
	if entering {
		if node.FirstChild() != nil {
			// Bugfix paragraph spacing.
			node.InsertBefore(node, node.FirstChild(), ast.NewString([]byte{}))
		}

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
				sourceFile := rundown.GetModifiers().GetValue("from")

				if sourceFile == nil {
					sourceFile = &r.ctx.CurrentFile
				} else {
					absPath := filepath.Join(filepath.Dir(r.ctx.CurrentFile), *sourceFile)
					sourceFile = &absPath
				}

				name := rundown.GetModifiers().GetValue("invoke")
				if name == nil {
					return ast.WalkStop, errors.New("Invoke requires a ShortCode value")
				}

				rd, err := LoadFile(*sourceFile)
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

					err := fmt.Sprintf("Cannot find %s in %s.", *name, *sourceFile)
					fns := rd.GetShortCodes().Functions

					if len(fns) > 0 {
						fnNames := []string{}
						for key := range fns {
							fnNames = append(fnNames, key)
						}

						err = fmt.Sprintf("%s. Valid functions: %s", err, strings.Join(fnNames, ", "))
					} else {
						err = fmt.Sprintf("%s. File not found, or has no functions.", err)
					}

					return ast.WalkStop, errors.New(err)
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
		// w.WriteString("\n")
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
	if entering {
		emojiNode := node.(*markdown.EmojiInline)
		w.WriteString(strings.TrimSpace(emoji.Sprint(":" + emojiNode.EmojiCode + ":")))
	}
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderTextSubInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Write([]byte("X"))
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderFixedText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *RundownRenderer) renderRundownInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	rd := node.(*markdown.RundownInline)

	if rd.Modifiers.Flags[markdown.Flag("sub-env")] && entering {
		// Find all text nodes containing $ENV or ${ENV:-} replace with FixedText.
		ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if text, ok := node.(*ast.Text); ok && entering {
				str := rdutil.NodeLines(text, source)
				newStr, err := SubEnv(str, r.ctx)

				if err == nil {
					if newStr != str {
						text.Parent().ReplaceChild(text.Parent(), text, ast.NewString([]byte(newStr)))
					}
				} else {
					text.Parent().InsertAfter(text.Parent(), text, ast.NewString([]byte(" (not set)")))
				}
			}

			return ast.WalkContinue, nil
		})

	}

	return ast.WalkContinue, nil
}

type FixedText struct {
	ast.BaseInline
	Contents string
}

var KindFixedText = ast.NewNodeKind("FixedText")

func (t *FixedText) Kind() ast.NodeKind {
	return KindFixedText
}

func (t *FixedText) Dump(source []byte, level int) {
	fmt.Printf("%sFixedText: %q\n", strings.Repeat("    ", level), t.Contents)
}

func NewFixedText(text string) *FixedText {

	return &FixedText{Contents: text}
}
