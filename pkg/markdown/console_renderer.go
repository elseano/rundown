package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	icolor "image/color"

	// "fmt"

	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	"github.com/alecthomas/chroma"
	formatters "github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"

	"github.com/logrusorgru/aurora"

	"github.com/eliukblau/pixterm/pkg/ansimage"

	rdutil "github.com/elseano/rundown/pkg/util"
)

type consoleRendererExt struct {
}

// Strikethrough is an extension that allow you to use invisibleBlock expression like '~~text~~' .
var ConsoleRenderer = &consoleRendererExt{}

func (e *consoleRendererExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewRenderer(), 1)))
}

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	TempDir        string
	RundownHandler RundownHandler
	Level          int
	ConsoleWidth   int
	LevelChange    func(level int)
}

const optTempDir renderer.OptionName = "TempDir"
const optRundownHandler renderer.OptionName = "RundownHandler"
const optLevelLevel renderer.OptionName = "LevelLevel"
const optConsoleWidth renderer.OptionName = "ConsoleWidth"
const optLevelChange renderer.OptionName = "LevelChange"

type ExecutionResult string

const (
	Continue ExecutionResult = "Continue"
	Skip                     = "Skip"
	Stop                     = "Stop"
)

type RundownHandler interface {
	Mutate([]byte, ast.Node) ([]byte, error)
	OnRundownNode(node ast.Node, entering bool) error
	OnExecute(node *ExecutionBlock, source []byte, out io.Writer) (ExecutionResult, error)
}

type withRundownHandler struct {
	handler RundownHandler
}

func (o *withRundownHandler) SetConfig(c *renderer.Config) {
	c.Options[optRundownHandler] = o.handler
}

func (o *withRundownHandler) SetConsoleOption(c *Config) {
	c.RundownHandler = o.handler
}

// The mutator is responsible for changing the inside of a RundownInline block before
// it's written to the output, but after it's been rendered.
func WithRundownHandler(handler RundownHandler) interface {
	renderer.Option
	Option
} {
	return &withRundownHandler{handler: handler}
}

type withLevel struct {
	level int
}

func (o *withLevel) SetConfig(c *renderer.Config) {
	c.Options[optLevelLevel] = o.level
}

func (o *withLevel) SetConsoleOption(c *Config) {
	c.Level = o.level
}

// The mutator is responsible for changing the inside of a RundownInline block before
// it's written to the output, but after it's been rendered.
func WithLevel(level int) interface {
	renderer.Option
	Option
} {
	return &withLevel{level: level}
}

type withLevelChange struct {
	changer func(level int)
}

func (o *withLevelChange) SetConfig(c *renderer.Config) {
	c.Options[optLevelChange] = o.changer
}

func (o *withLevelChange) SetConsoleOption(c *Config) {
	c.LevelChange = o.changer
}

// The mutator is responsible for changing the inside of a RundownInline block before
// it's written to the output, but after it's been rendered.
func WithLevelChange(changer func(level int)) interface {
	renderer.Option
	Option
} {
	return &withLevelChange{changer: changer}
}

type withConsoleWidth struct {
	width int
}

func (o *withConsoleWidth) SetConfig(c *renderer.Config) {
	c.Options[optConsoleWidth] = o.width
}

func (o *withConsoleWidth) SetConsoleOption(c *Config) {
	c.ConsoleWidth = o.width
}

// The mutator is responsible for changing the inside of a RundownInline block before
// it's written to the output, but after it's been rendered.
func WithConsoleWidth(width int) interface {
	renderer.Option
	Option
} {
	return &withConsoleWidth{width: width}
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	tmpDir, err := ioutil.TempDir("", "rundown")

	if err != nil {
		panic(err)
	}

	return Config{
		TempDir:      tmpDir,
		ConsoleWidth: 80,
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optTempDir:
		c.TempDir = value.(string)
	case optRundownHandler:
		c.RundownHandler = value.(RundownHandler)
	case optLevelLevel:
		c.Level = value.(int)
	case optConsoleWidth:
		c.ConsoleWidth = value.(int)
	case optLevelChange:
		c.LevelChange = value.(func(level int))
		// case optXHTML:
		// 	c.XHTML = value.(bool)
		// case optUnsafe:
		// 	c.Unsafe = value.(bool)
		// case optTextWriter:
		// 	c.Writer = value.(Writer)
	}
}

// An Option interface sets options for HTML based renderers.
type Option interface {
	SetConsoleOption(*Config)
}

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as Console strings
type Renderer struct {
	Config
	blockStyles       *StyleStack
	inlineStyles      *StyleStack
	currentLevel      int
	currentlySkipping bool
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config:            NewConfig(),
		blockStyles:       NewStyleStack(),
		inlineStyles:      NewStyleStack(),
		currentLevel:      1,
		currentlySkipping: false,
	}

	for _, opt := range opts {
		opt.SetConsoleOption(&r.Config)
	}

	r.SetLevel(r.Config.Level)

	// r.currentLevel = r.Config.LevelLevel

	return r
}

func NewStyleStack() *StyleStack {
	return &StyleStack{
		Styles: []Style{},
	}
}

type StyleStack struct {
	Styles []Style
}

func (s *StyleStack) Push(style Style) {
	s.Styles = append(s.Styles, style)
}

func (s *StyleStack) Pop() Style {
	if len(s.Styles) == 0 {
		return nil
	}

	style := s.Styles[len(s.Styles)-1]
	if len(s.Styles) > 1 {
		s.Styles = s.Styles[0 : len(s.Styles)-1]
	} else {
		s.Styles = []Style{}
	}

	return style
}

func (s *StyleStack) Peek() Style {
	if len(s.Styles) == 0 {
		return nil
	}

	return s.Styles[len(s.Styles)-1]
}

type Style interface {
	Wrap(str string) string
	Begin() string
	End() string
}

type Color aurora.Color

const (
	resetStyle Color = 0
)

func (c Color) Begin() string {
	return "\033[" + aurora.Color(c).Nos(false) + "m"
}

func (c Color) End() string {
	return "\033[0m"
}

func (c Color) Wrap(str string) string {
	return c.Begin() + str + c.End()
}

func (r *Renderer) CurrentLevel() int {
	return r.currentLevel
}

func (r *Renderer) SetLevel(level int) {
	// Make No Heading & First Heading at the same level.
	if level < 1 {
		level = 1
	}
	r.currentLevel = level

	if r.Config.LevelChange != nil {
		r.Config.LevelChange(level)
	}
}

func (r *Renderer) nodeLinesToString(source []byte, n ast.Node) string {
	var buffer bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		buffer.Write(line.Value(source))
	}

	return buffer.String()
}

func (r *Renderer) levelLines(lines string, b util.BufWriter) {
	splitlines := strings.Split(lines, "\n")
	for _, v := range splitlines {
		r.writeString(b, paddingForLevel(r.currentLevel)+v+"\n")
	}
}

func (r *Renderer) levelLinesWithPrefix(prefix string, lines string, b util.BufWriter) {
	splitlines := strings.Split(lines, "\n")
	for _, v := range splitlines {
		r.writeString(b, paddingForLevel(r.currentLevel)+prefix+v+"\n")
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.skippable(r.renderDocument))
	reg.Register(ast.KindHeading, r.skippable(r.renderHeading))
	reg.Register(ast.KindBlockquote, r.skippable(r.renderBlockquote))
	reg.Register(ast.KindCodeBlock, r.skippable(r.renderCodeBlock))
	reg.Register(ast.KindFencedCodeBlock, r.skippable(r.renderFencedCodeBlock))
	reg.Register(ast.KindHTMLBlock, r.skippable(r.renderHTMLBlock))
	reg.Register(ast.KindList, r.skippable(r.renderList))
	reg.Register(ast.KindListItem, r.skippable(r.renderListItem))
	reg.Register(ast.KindParagraph, r.skippable(r.renderParagraph))
	reg.Register(ast.KindTextBlock, r.skippable(r.renderTextBlock))
	reg.Register(ast.KindThematicBreak, r.skippable(r.renderThematicBreak))
	reg.Register(KindRundownBlock, r.skippable(r.renderRundownBlock))
	reg.Register(KindSection, r.skippable(r.renderNothing))
	reg.Register(KindSectionedDocument, r.skippable(r.renderNothing))
	reg.Register(KindContainer, r.skippable(r.renderNothing))
	reg.Register(KindExecutionBlock, r.skippable(r.renderExecutionBlock))

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(extast.KindStrikethrough, r.renderStrikethrough)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderNothing)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
	reg.Register(KindRundownInline, r.renderRundownInline)
	reg.Register(KindCodeModifierBlock, r.renderNothing)
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		_, _ = w.Write(line.Value(source))
	}
}

func (r *Renderer) skippable(render func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error)) func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if _, isSection := node.(*Section); isSection {
			r.currentlySkipping = false
		}

		if r.currentlySkipping {
			return ast.WalkSkipChildren, nil
		}

		return render(w, source, node, entering)
	}
}

// GlobalAttributeFilter defines attribute names which any elements can have.
var GlobalAttributeFilter = util.NewBytesFilter(
	[]byte("accesskey"),
	[]byte("autocapitalize"),
	[]byte("class"),
	[]byte("contenteditable"),
	[]byte("contextmenu"),
	[]byte("dir"),
	[]byte("draggable"),
	[]byte("dropzone"),
	[]byte("hidden"),
	[]byte("id"),
	[]byte("itemprop"),
	[]byte("lang"),
	[]byte("slot"),
	[]byte("spellcheck"),
	[]byte("style"),
	[]byte("tabindex"),
	[]byte("title"),
	[]byte("translate"),
)

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.SetLevel(r.Config.Level)

	return ast.WalkContinue, nil
}

func (r *Renderer) renderNothing(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkContinue, nil
}

func (r *Renderer) renderRundownBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	rundown := node.(*RundownBlock)

	if r.Config.RundownHandler != nil {
		err := r.Config.RundownHandler.OnRundownNode(node, entering)
		if err != nil {
			return ast.WalkStop, err
		}
	}

	// We don't render function/shortcode option contents, thats for reading only.
	if rundown.Modifiers.HasAny("ignore", "opt") {
		return ast.WalkSkipChildren, nil
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderExecutionBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if r.Config.RundownHandler != nil {
			result, err := r.Config.RundownHandler.OnExecute(node.(*ExecutionBlock), source, w)
			if err != nil {
				return ast.WalkStop, err
			}

			switch result {
			case Continue:
				return ast.WalkContinue, nil
			case Skip:
				r.currentlySkipping = true
				return ast.WalkSkipChildren, nil
			case Stop:
				return ast.WalkStop, nil
			}
		}
	} else if _, ok := node.NextSibling().(*ExecutionBlock); !ok {
		_, _ = w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderRundownInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	var buf bytes.Buffer
	var w2 = bufio.NewWriter(&buf)

	if !node.HasChildren() {
		return ast.WalkSkipChildren, nil
	}

	if entering {

		renderer := PrepareMarkdown().Renderer()
		// TODO - Copy all config and apply it to the renderer.

		rundown := node.(*RundownInline)

		if rundown.Modifiers.Flags[Flag("ignore")] == true {
			return ast.WalkSkipChildren, nil
		}

		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			renderer.Render(w2, source, child)
		}

		if r.Config.RundownHandler != nil {
			err := r.Config.RundownHandler.OnRundownNode(node, entering)
			if err != nil {
				return ast.WalkStop, err
			}

			s := string(buf.Bytes())
			result, err := r.Config.RundownHandler.Mutate([]byte(s), node)
			if err != nil {
				return ast.WalkStop, err
			}
			w.Write(result)
		} else {
			w.Write(buf.Bytes())
		}
	} else {
		if r.Config.RundownHandler != nil {
			err := r.Config.RundownHandler.OnRundownNode(node, entering)
			if err != nil {
				return ast.WalkStop, err
			}
		}
	}

	return ast.WalkSkipChildren, nil
}

// HeadingAttributeFilter defines attribute names which heading elements can have
var HeadingAttributeFilter = GlobalAttributeFilter

func paddingForLevel(level int) string {
	level--
	if level < 0 {
		level = 0
	}

	return strings.Repeat("  ", level)
}

func (r *Renderer) writeString(w util.BufWriter, s string) {
	_, _ = w.WriteString(s)
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		if node.PreviousSibling() != nil {
			// r.writeString(w, "\n")
		}

		r.writeString(w, paddingForLevel(n.Level))

		switch n.Level {
		case 1:
			r.inlineStyles.Push(Color(aurora.CyanFg | aurora.BoldFm | aurora.UnderlineFm))
		case 2:
			r.inlineStyles.Push(Color(aurora.CyanFg | aurora.BoldFm))
		case 3:
			r.inlineStyles.Push(Color(aurora.CyanFg))
		default:
			r.inlineStyles.Push(Color(aurora.BoldFm))
		}

		r.SetLevel(n.Level)
	} else {
		r.inlineStyles.Pop()

		mods := NewModifiers()

		if r, ok := node.PreviousSibling().(*RundownBlock); ok {
			mods.Ingest(r.Modifiers)
		}

		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if r, ok := c.(*RundownInline); ok {
				mods.Ingest(r.Modifiers)
			}
		}

		if label, ok := mods.Values[Parameter("label")]; ok {
			r.writeString(w, "\033[0m")
			r.writeString(w, aurora.Faint(" ("+label+")").String())
		}

		r.writeString(w, "\n")
	}
	return ast.WalkContinue, nil
}

// BlockquoteAttributeFilter defines attribute names which blockquote elements can have
var BlockquoteAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("cite"),
)

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.blockStyles.Push(NewBulletSequence(aurora.Blue(" >").String(), paddingForLevel(r.currentLevel)))
	} else {
		r.blockStyles.Pop()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) syntaxHighlightText(w util.BufWriter, language string, source []byte, n ast.Node) {
	lang := lexers.Get(language)
	target := r.nodeLinesToString(source, n)
	var buf bytes.Buffer

	if lang == nil {
		lang = lexers.Analyse(string(source))
	}

	if lang != nil {
		lexer := chroma.Coalesce(lexers.Get(language))
		formatter := formatters.TTY256

		iterator, _ := lexer.Tokenise(nil, target)
		_ = formatter.Format(&buf, styles.Pygments, iterator) == nil
	} else {
		buf.WriteString(target)
	}

	// Trim any trailing formatting only lines.
	trailing := []string{}
	lines := strings.SplitAfter(buf.String(), "\n")
	for i := len(lines) - 1; i >= 0 && strings.TrimSpace(rdutil.RemoveColors(lines[i])) == ""; {
		trailing = append(trailing, strings.TrimSpace(lines[i]))
		lines = lines[0:i]
		i = i - 1
	}

	r.levelLinesWithPrefix(aurora.Black(" ┃ ").Faint().String(), strings.TrimSpace(strings.Join(lines, "")), w)
	for i := len(trailing) - 1; i >= 0; i = i - 1 {
		w.WriteString(trailing[i])
	}

}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.levelLinesWithPrefix(aurora.Black(" ┃ ").Faint().String(), strings.TrimSpace(r.nodeLinesToString(source, n)), w)
	} else {
		_, _ = w.WriteString("\r\n") // Block level element, add blank line.
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		language := n.Language(source)
		if language != nil {
			r.syntaxHighlightText(w, string(language), source, node)
		}
	} else {
		_, _ = w.WriteString("\r\n") // Block level element, add blank line.
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// We dont render anything
	return ast.WalkContinue, nil
}

// ListAttributeFilter defines attribute names which list elements can have.
var ListAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("start"),
	[]byte("reversed"),
)

type IntegerSequenceStyle struct {
	seq   int
	level string
}

func NewIntegerSequence(start int, level string) *IntegerSequenceStyle {
	return &IntegerSequenceStyle{seq: start - 1, level: level}
}

func (s *IntegerSequenceStyle) Begin() string {
	s.seq++
	return s.level + aurora.Blue(strconv.Itoa(s.seq)+" ").Bold().String()
}

func (s *IntegerSequenceStyle) End() string {
	return ""
}

func (s *IntegerSequenceStyle) Wrap(str string) string {
	return s.Begin() + str + s.End()
}

type BulletSequenceStyle struct {
	marker string
}

func NewBulletSequence(marker string, level string) BulletSequenceStyle {
	return BulletSequenceStyle{marker: level + marker}
}

func (s BulletSequenceStyle) Begin() string {
	return aurora.Bold(s.marker + " ").String()
}

func (s BulletSequenceStyle) End() string {
	return ""
}

func (s BulletSequenceStyle) Wrap(str string) string {
	return s.Begin() + str + s.End()
}

var bulletLevels = []string{"•", "◦", "⁃"}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)

	if entering {
		if n.IsOrdered() {
			r.blockStyles.Push(NewIntegerSequence(n.Start, paddingForLevel(r.currentLevel)))
		} else {
			var (
				depth          = 0
				p     ast.Node = n
			)

			for {
				if _, ok := p.Parent().(*ast.ListItem); ok {
					depth++
					p = p.Parent()
				} else {
					break
				}

				if depth > 2 {
					depth = 2
					break
				}
			}

			r.blockStyles.Push(NewBulletSequence(bulletLevels[depth], paddingForLevel(r.currentLevel)))
		}
		r.SetLevel(r.currentLevel + 1)
	} else {
		// if _, ok := node.Parent().(*ast.ListItem); !ok {
		_, _ = w.WriteString("\n")
		// }
		r.blockStyles.Pop()
		r.SetLevel(r.currentLevel - 1)
	}
	return ast.WalkContinue, nil
}

// ListItemAttributeFilter defines attribute names which list item elements can have.
var ListItemAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("value"),
)

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// fc := n.FirstChild()
		// if fc != nil {
		// 	if _, ok := fc.(*ast.TextBlock); !ok {
		// 		_ = w.WriteByte('\n')
		// 	}
		// }

		_, _ = w.WriteString(r.blockStyles.Peek().Begin())

	} else {
		bs := r.blockStyles.Peek()
		if bs == nil {
			fmt.Print("ERROR NO STYLE\n")
			n.Dump(source, 1)
		} else {
			_, _ = w.WriteString(r.blockStyles.Peek().End())
		}

		// If we have a List as the last child, then that list will have
		// already placed us on a newline. Otherwise, print a newline.
		if _, isList := n.LastChild().(*ast.List); !isList {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

// ParagraphAttributeFilter defines attribute names which paragraph elements can have.
var ParagraphAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		link, ok := n.FirstChild().(*ast.Link)

		if ok && n.ChildCount() == 1 && link.Title == nil {
			// Modifier
			return ast.WalkSkipChildren, nil
		}

		// Indent in two cases:
		// 1. We're not inside a list.
		// 2. We're inside a list, but not the first child.
		if _, ok := n.Parent().(*ast.ListItem); !ok || n.Parent().FirstChild() != n {
			r.writeString(w, paddingForLevel(r.currentLevel))
		}

		// if n.Attributes() != nil {
		// 	_, _ = w.WriteString("<p")
		// 	RenderAttributes(w, n, ParagraphAttributeFilter)
		// 	_ = w.WriteByte('>')
		// } else {
		// 	_, _ = w.WriteString("<p>")
		// }
	} else {
		// Loose lists end with with ListItem > Paragraph, which breaks formatting
		// on the console. So don't insert newlines.
		if _, ok := n.Parent().(*ast.ListItem); !ok {
			_, _ = w.WriteString("\n\n")
		} else if n.NextSibling() != nil {
			// However, if there's additional block elements after the paragraph,
			// then we'll make sure that block element is on it's own line.
			_, _ = w.WriteString("\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {

		// Special case, we have an embedded list as the next sibling, add a newline.
		if _, ok := n.Parent().(*ast.ListItem); ok {
			if _, isList := n.NextSibling().(*ast.List); isList {
				_ = w.WriteByte('\n')
			}
		}

		// Otherwise, just let the parent ListItem handle the line breaks.
	}
	return ast.WalkContinue, nil
}

// ThematicAttributeFilter defines attribute names which hr elements can have.
var ThematicAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("align"),   // [Deprecated]
	[]byte("color"),   // [Not Standardized]
	[]byte("noshade"), // [Deprecated]
	[]byte("size"),    // [Deprecated]
	[]byte("width"),   // [Deprecated]
)

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	line := strings.Repeat("-", r.Config.ConsoleWidth-4)

	_, _ = w.WriteString(fmt.Sprintf("  %s  \r\n\r\n", aurora.Faint(line).String()))

	return ast.WalkContinue, nil
}

// LinkAttributeFilter defines attribute names which link elements can have.
var LinkAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("download"),
	// []byte("href"),
	[]byte("hreflang"),
	[]byte("media"),
	[]byte("ping"),
	[]byte("referrerpolicy"),
	[]byte("rel"),
	[]byte("shape"),
	[]byte("target"),
)

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<a href="`)
	url := n.URL(source)
	label := n.Label(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(url), []byte("mailto:")) {
		_, _ = w.WriteString("mailto:")
	}
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(url, false)))
	if n.Attributes() != nil {
		_ = w.WriteByte('"')
		RenderAttributes(w, n, LinkAttributeFilter)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString(`">`)
	}
	_, _ = w.Write(util.EscapeHTML(label))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

// CodeAttributeFilter defines attribute names which code elements can have.
var CodeAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {

		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)

			// Swallows newline, changing it to a space.
			if bytes.HasSuffix(value, []byte("\n")) {

				_, _ = w.WriteString(aurora.Yellow(string(value[:len(value)-1])).String())
				if c != n.LastChild() {
					_, _ = w.Write([]byte(" "))
				}
			} else {
				_, _ = w.WriteString(aurora.Yellow(string(value)).String())
				// r.Writer.RawWrite(w, value)
			}
		}

		return ast.WalkSkipChildren, nil
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrikethrough(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.inlineStyles.Push(Color(aurora.StrikeThroughFm | aurora.BlackFg))
	} else {
		r.inlineStyles.Pop()
	}

	return ast.WalkContinue, nil
}

// EmphasisAttributeFilter defines attribute names which emphasis elements can have.
var EmphasisAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	if entering {
		if n.Level == 2 {
			w.WriteString(Color(aurora.BoldFm).Begin())
		} else {
			w.WriteString(Color(aurora.ItalicFm).Begin())
		}
	} else {
		if n.Level == 2 {
			w.WriteString(Color(aurora.BoldFm).End())
		} else {
			w.WriteString(Color(aurora.ItalicFm).End())
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		r.inlineStyles.Push(Color(aurora.UnderlineFm))
		w.WriteString("\033]8;;" + string(n.Destination) + "\033\\")
	} else {
		r.inlineStyles.Pop()
		w.WriteString("\033]8;;\033\\")

		// if n.Title != nil {
		// 	_, _ = w.WriteString(aurora.Faint(" (" + string(n.Title) + ")").String())
		// } else {
		// 	_, _ = w.WriteString(aurora.Faint(" (").String() + aurora.Faint(string(n.Destination)).String() + aurora.Faint(")").String())
		// }
	}

	return ast.WalkContinue, nil
}

// ImageAttributeFilter defines attribute names which image elements can have.
var ImageAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("align"),
	[]byte("border"),
	[]byte("crossorigin"),
	[]byte("decoding"),
	[]byte("height"),
	[]byte("importance"),
	[]byte("intrinsicsize"),
	[]byte("ismap"),
	[]byte("loading"),
	[]byte("referrerpolicy"),
	[]byte("sizes"),
	[]byte("srcset"),
	[]byte("usemap"),
	[]byte("width"),
)

var urlMatch = regexp.MustCompile("^http(s?)://")

func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	if val, set := os.LookupEnv("COLORTERM"); !set || (val != "truecolor" && val != "24bit") {
		// Only render images when we're sure we're in truecolor mode.
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)
	var image *ansimage.ANSImage
	var maxWidth = r.Config.ConsoleWidth - 1 - (r.currentLevel * 2)

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
		lines := strings.Split(image.Render(), "\n")
		for _, line := range lines {
			w.WriteString(line + "\n")

			// Must flush every line to avoid filling buffer.
			// Otherwise WordWrapWriter gets confused by half-formed escape codes on buffer full.
			w.Flush()
		}
	}

	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment

	if style := r.inlineStyles.Peek(); style != nil {
		r.writeString(w, style.Begin())
	}

	value := segment.Value(source)

	// If the next node is a rundown marker indicating a label, trim
	// any trailing space to present the label properly.
	if rundown, ok := node.NextSibling().(*RundownInline); ok {
		if _, ok := rundown.Modifiers.Values["label"]; ok {
			value = bytes.TrimRight(value, " ")
		}
	}

	if n.IsRaw() {
		_, _ = w.Write(value)
	} else {
		_, _ = w.Write(value)
		if n.HardLineBreak() || (n.SoftLineBreak() && true) {
			_, _ = w.WriteString("\n")
			r.writeString(w, paddingForLevel(r.currentLevel))
		} else if n.SoftLineBreak() {
			_ = w.WriteByte('\n')
			r.writeString(w, paddingForLevel(r.currentLevel))
		}
	}

	if style := r.inlineStyles.Peek(); style != nil {
		r.writeString(w, style.End())
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	if n.IsCode() {
		_, _ = w.Write(n.Value)
	} else {
		// if n.IsRaw() {
		w.Write(n.Value)
		// } else {
		// 	r.Writer.Write(w, n.Value)
		// }
	}
	return ast.WalkContinue, nil
}

var dataPrefix = []byte("data-")

// RenderAttributes renders given node's attributes.
// You can specify attribute names to render by the filter.
// If filter is nil, RenderAttributes renders all attributes.
func RenderAttributes(w util.BufWriter, node ast.Node, filter util.BytesFilter) {
	for _, attr := range node.Attributes() {
		if filter != nil && !filter.Contains(attr.Name) {
			if !bytes.HasPrefix(attr.Name, dataPrefix) {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		// TODO: convert numeric values to strings
		_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
		_ = w.WriteByte('"')
	}
}

// A Writer interface writes textual contents to a writer.
type Writer interface {
	// Write writes the given source to writer with resolving references and unescaping
	// backslash escaped characters.
	Write(writer util.BufWriter, source []byte)

	// RawWrite writes the given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWrite(writer util.BufWriter, source []byte)
}

type defaultWriter struct {
}

// DefaultWriter is a default implementation of the Writer.
var DefaultWriter = &defaultWriter{}
