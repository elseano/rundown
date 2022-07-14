package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"time"

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
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"

	"github.com/logrusorgru/aurora"

	"github.com/eliukblau/pixterm/pkg/ansimage"

	rundown_ast "github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/errs"
	"github.com/elseano/rundown/pkg/exec"
	rundown_renderer "github.com/elseano/rundown/pkg/renderer"
	"github.com/elseano/rundown/pkg/renderer/term/spinner"
	"github.com/elseano/rundown/pkg/text"
	emoji_ast "github.com/yuin/goldmark-emoji/ast"

	rundown_styles "github.com/elseano/rundown/pkg/renderer/styles"
	rdutil "github.com/elseano/rundown/pkg/util"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
)

type consoleRendererExt struct {
}

// Strikethrough is an extension that allow you to use invisibleBlock expression like '~~text~~' .
var ConsoleRenderer = &consoleRendererExt{}

func (e *consoleRendererExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewRenderer(rundown_renderer.NewContext("<>")), 1)))
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
	OnRundownNode(node ast.Node, entering bool) (ast.WalkStatus, error)
	OnExecute(node *rundown_ast.ExecutionBlock, source []byte, out util.BufWriter) (ExecutionResult, error)
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
	Context           *rundown_renderer.Context
	exitCode          int
	skipUntil         ast.Node
	wrappingWriter    *wordwrap.WordWrap
	nonWrappingWriter util.BufWriter
	lastRendered      ast.Node
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(context *rundown_renderer.Context, opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config:            NewConfig(),
		blockStyles:       NewStyleStack(),
		inlineStyles:      NewStyleStack(),
		currentLevel:      1,
		currentlySkipping: false,
		Context:           context,
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

func (s *StyleStack) Push(style Style) Style {
	s.Styles = append(s.Styles, style)

	return style
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
	if ColorsEnabled {
		return "\033[" + aurora.Color(c).Nos(false) + "m"
	} else {
		return ""
	}
}

func (c Color) End() string {
	if ColorsEnabled {
		return "\033[0m"
	} else {
		return ""
	}
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

func (r *Renderer) writeLinesWithPrefix(prefix string, lines string, b util.BufWriter) {
	splitlines := strings.Split(lines, "\n")
	for _, v := range splitlines {
		r.writeString(b, prefix+v+"\n")
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.supportSkipping(r.blockCommon(r.renderDocument)))
	reg.Register(ast.KindHeading, r.supportSkipping(r.blockCommon(r.renderHeading)))
	reg.Register(ast.KindBlockquote, r.supportSkipping(r.blockCommon(r.renderBlockquote)))
	reg.Register(ast.KindCodeBlock, r.supportSkipping(r.blockCommon(r.renderCodeBlock)))
	reg.Register(ast.KindFencedCodeBlock, r.supportSkipping(r.blockCommon(r.renderFencedCodeBlock)))
	reg.Register(ast.KindHTMLBlock, r.supportSkipping(r.blockCommon(r.renderHTMLBlock)))
	reg.Register(ast.KindList, r.supportSkipping(r.blockCommon(r.renderList)))
	reg.Register(ast.KindListItem, r.supportSkipping(r.blockCommon(r.renderListItem)))
	reg.Register(ast.KindParagraph, r.wrapBlock(r.supportSkipping(r.blockCommon(r.renderParagraph))))
	reg.Register(ast.KindTextBlock, r.supportSkipping(r.blockCommon(r.renderTextBlock)))
	reg.Register(ast.KindThematicBreak, r.supportSkipping(r.blockCommon(r.renderThematicBreak)))
	// reg.Register(rundown_ast.KindRundownBlock, r.blockCommon(r.renderRundownBlock))
	// reg.Register(KindSection, r.blockCommon(r.renderNothing))
	reg.Register(rundown_ast.KindExecutionBlock, r.blockCommon(r.renderExecutionBlock))

	// inlines

	reg.Register(ast.KindAutoLink, r.wrapInline(r.renderAutoLink))
	reg.Register(ast.KindCodeSpan, r.wrapInline(r.renderCodeSpan))
	reg.Register(extast.KindStrikethrough, r.wrapInline(r.renderStrikethrough))
	reg.Register(ast.KindEmphasis, r.wrapInline(r.renderEmphasis))
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.wrapInline(r.renderLink))
	reg.Register(ast.KindRawHTML, r.renderHollow)
	reg.Register(ast.KindText, r.wrapInline(r.renderText))
	reg.Register(ast.KindString, r.wrapInline(r.renderString))

	reg.Register(emoji_ast.KindEmoji, r.wrapInline(r.renderEmoji))

	// other
	reg.Register(rundown_ast.KindDescriptionBlock, r.supportSkipping(r.renderHollow))
	reg.Register(rundown_ast.KindEnvironmentSubstitution, r.wrapInline(r.supportSkipping(r.renderEnvironmentSubstitution)))
	// reg.Register(rundown_ast.KindExecutionBlock, r.renderTodo))
	reg.Register(rundown_ast.KindIgnoreBlock, r.supportSkipping(r.renderTodo("Ignore")))
	reg.Register(rundown_ast.KindOnFailure, r.supportSkipping(r.renderNothing))
	reg.Register(rundown_ast.KindRundownBlock, r.supportSkipping(r.renderTodo("Rundown")))
	reg.Register(rundown_ast.KindSaveCodeBlock, r.supportSkipping(r.renderSaveCodeBlock))
	reg.Register(rundown_ast.KindSectionOption, r.supportSkipping(r.renderHollow))
	reg.Register(rundown_ast.KindSectionPointer, r.supportSkipping(r.renderHollow))
	reg.Register(rundown_ast.KindStopFail, r.supportSkipping(r.renderStopFail))
	reg.Register(rundown_ast.KindStopOk, r.supportSkipping(r.renderStopOk))
	reg.Register(rundown_ast.KindSubEnvBlock, r.supportSkipping(r.renderHollow))
	reg.Register(rundown_ast.KindInvokeBlock, r.supportSkipping(r.renderInvokeBlock))
	reg.Register(rundown_ast.KindSkipBlock, r.renderSkipBlock)

	// Conditional blocks are transparent, they shouldn't render.
	reg.Register(rundown_ast.KindConditionalStart, r.supportSkipping(r.renderHollow))
	reg.Register(rundown_ast.KindConditionalEnd, r.supportSkipping(r.renderHollow))

	// reg.Register(KindRundownInline, r.renderRundownInline)
	// reg.Register(KindCodeModifierBlock, r.renderNothing)
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		_, _ = w.Write(line.Value(source))
	}
}

func (r *Renderer) blockCommon(render func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error)) func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		result, err := render(w, source, node, entering)
		w.Flush()
		return result, err
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

	if !entering {
		r.ensureBlockSeparator(w, node)

		if r.exitCode != 0 {
			return ast.WalkStop, fmt.Errorf("")
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) StartAt(node ast.Node) {
	r.skipUntil = node
}

func (r *Renderer) renderHollow(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderNothing(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderInvokeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	invoke := node.(*rundown_ast.InvokeBlock)

	if entering {
		// If this is a dependency invoke, and it's already been run, then skip it.
		if complete, present := r.Context.DepsCompleted[invoke.Invoke]; complete && present && invoke.AsDependency {
			return ast.WalkSkipChildren, nil
		}

		// Otherwise, snapshot the environment, and reset it for the invoked code.
		invoke.PreviousEnv = r.Context.Env
		r.Context.ResetEnv()

		section := invoke.Target
		if section == nil {
			return ast.WalkStop, fmt.Errorf("invalid section: %s", invoke.Invoke)
		}

		rdutil.Logger.Debug().Msgf("Invoking %s with: %+v", invoke.Invoke, invoke.Args)

		env, err := section.ParseOptionsWithResolution(invoke.Args, r.Context.Env)
		if err != nil {
			return ast.WalkStop, err
		}

		rdutil.Logger.Debug().Msgf("Parsed options are: %#v", env)

		r.Context.ImportEnv(env)
	} else {
		r.Context.ResetEnv()
		r.Context.ImportEnv(invoke.PreviousEnv)
		invoke.PreviousEnv = nil

		if invoke.AsDependency {
			r.Context.DepsCompleted[invoke.Invoke] = true
		}
	}

	return ast.WalkContinue, nil
}

func runIfScript(ctx *rundown_renderer.Context, ifScript string) (bool, error) {
	// Allow unset variables here, typically the script will be checking for these.
	ifScript = fmt.Sprintf("set +u\n%s\n", ifScript)

	runner := exec.NewRunner()
	runner.ImportEnv(ctx.Env)

	outputBuffer := bytes.Buffer{}
	outputWait := sync.WaitGroup{}

	_, err := runner.SetScript("sh", "sh", []byte(ifScript))

	if err != nil {
		return false, err
	}

	process, err := runner.Prepare()
	if err != nil {
		return false, err
	}

	outputWait.Add(1)
	go func() {
		io.Copy(&outputBuffer, process.Stderr)
		outputWait.Done()
	}()

	err = process.Start()

	if err != nil {
		return false, err
	}

	outputWait.Wait()
	exitCode, _, err := process.Wait()

	rdutil.Logger.Debug().Msgf("Process output: %s", outputBuffer.String())

	if err != nil {
		rdutil.Logger.Debug().Msgf("Error: %#v", err)

		return false, err
	}

	rdutil.Logger.Debug().Msgf("If Script Result: %d", exitCode)

	return exitCode == 0, err
}

func (r *Renderer) renderStopFail(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		r.inlineStyles.Pop()
		w.WriteString("\n")
		if r.exitCode != 0 {
			return ast.WalkStop, &errs.ExecutionError{ExitCode: r.exitCode}
		}
		w.Flush()
		return ast.WalkStop, errs.ErrStopFail
	} else {
		r.inlineStyles.Push(Color(aurora.RedFg))

		if node.ChildCount() > 0 {
			w.WriteString(aurora.Red("✖ ").String())
		}

		return ast.WalkContinue, nil
	}
}

func (r *Renderer) renderStopOk(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		w.WriteString("\n")
		return ast.WalkStop, nil
	} else {
		return ast.WalkContinue, nil
	}
}

func (r *Renderer) renderEmoji(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		emoji := node.(*emoji_ast.Emoji)
		w.WriteString(string(emoji.Value.Unicode))
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderTodo(message string) func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			w.WriteString(fmt.Sprintf("[TODO - %s]", message))
		}

		return ast.WalkContinue, nil
	}
}

func (r *Renderer) renderRundownBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	rundown := node.(*rundown_ast.RundownBlock)

	if r.Config.RundownHandler != nil {
		cmd, err := r.Config.RundownHandler.OnRundownNode(rundown, entering)
		if err != nil {
			return ast.WalkStop, err
		}

		return cmd, nil
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderSaveCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	scb := node.(*rundown_ast.SaveCodeBlock)

	reader := text.NewNodeReaderFromSource(scb.CodeBlock, source)
	contents, _ := ioutil.ReadAll(reader)

	rdutil.Logger.Debug().Msgf("Rendering SaveCodeBlock...")

	rdutil.Logger.Debug().Msgf("Env is currently: %#v", r.Context.Env)

	for k, v := range scb.Replacements {
		if strings.HasPrefix(v, "$") {
			v = r.Context.Env[v[1:]]
		}

		rdutil.Logger.Debug().Msgf("Replacing %s with %s...", k, v)
		contents = bytes.Replace(contents, []byte(k), []byte(v), -1)
	}

	rdutil.Logger.Debug().Msgf("Contents: %s", string(contents))

	newFile, err := r.Context.CreateTempFile(scb.SaveToVariable)
	if err != nil {
		return ast.WalkStop, err
	}

	defer newFile.Close()

	_, err = newFile.Write(contents)
	if err != nil {
		return ast.WalkStop, err
	}

	if scb.Reveal {
		lang := string(scb.CodeBlock.Info.Text(source))
		r.syntaxHighlightBytes(w, lang, string(contents), false)
		w.WriteString("\n")
	}

	return ast.WalkContinue, nil
}

func friendlyDuration(d time.Duration) string {
	return d.String()
}

// Creates the correct spinner based on the environment.
func createSpinner(writer io.Writer, env map[string]string) Spinner {
	var s Spinner

	// Ensure we always immediately flush when writing the spinner.
	writer = NewFlushingWriter(writer)

	if NewSpinnerFunc != nil {
		s = NewSpinnerFunc(writer)
	} else if _, gitlab := os.LookupEnv("GITLAB_CI"); gitlab {
		s = spinner.NewGitlabSpinner(writer, Aurora)
	} else if _, ci := os.LookupEnv("CI"); ci {
		s = spinner.NewCISpinner(writer, Aurora)
	} else {
		s = spinner.NewStdoutSpinner(Aurora, ColorsEnabled, writer)
	}

	return spinner.NewSubenvSpinner(env, s)
}

func (r *Renderer) renderExecutionBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Check if we're currently skipping
	executionBlock := node.(*rundown_ast.ExecutionBlock)

	if r.skipUntil != nil {
		if executionBlock != r.skipUntil {
			rdutil.Logger.Debug().Msgf("Skipping %T", executionBlock)
			return ast.WalkSkipChildren, nil
		} else {
			rdutil.Logger.Debug().Msgf("Reached skip target. Rendering this one %T", executionBlock)
			r.skipUntil = nil
		}
	}

	if entering {
		return ast.WalkContinue, nil
	}

	if !executionBlock.Execute {
		return ast.WalkContinue, nil
	}

	contentReader := text.NewNodeReaderFromSource(executionBlock.CodeBlock, source)

	scriptContents, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return ast.WalkStop, err
	}

	rdutil.Logger.Debug().Msgf("Command to run script is: %s", executionBlock.With)
	rdutil.Logger.Debug().Msgf("Script is: %s", executionBlock.With)

	runner := exec.NewRunner()
	script, err := runner.SetScript(executionBlock.With, executionBlock.Language, scriptContents)
	if err != nil {
		return ast.WalkStop, err
	}

	runner.ImportEnv(r.Context.Env)

	/***** SPINNERS *****/
	var theSpinner Spinner

	rdutil.Logger.Debug().Msgf("Spinner mode: %d", executionBlock.SpinnerMode)

	switch executionBlock.SpinnerMode {
	case rundown_ast.SpinnerModeHidden:
		theSpinner = spinner.NewNullSpinner()
	case rundown_ast.SpinnerModeVisible:
		theSpinner = createSpinner(w, r.Context.Env)
	case rundown_ast.SpinnerModeInlineAll:
		theSpinner = createSpinner(w, r.Context.Env)
		rdutil.Logger.Debug().Msgf("Stepped spinners.")
		script.Contents = exec.ChangeCommentsToSpinnerCommands(executionBlock.Language, script.Contents)
	}

	/***** BORG MODE *****/
	if executionBlock.ReplaceProcess {
		return ast.WalkStop, runner.RunReplacingProcess()
	}

	theSpinner.SetMessage(executionBlock.SpinnerName)
	theSpinner.Start()

	ifResult, err := r.checkIfScript(executionBlock)

	if err != nil {
		theSpinner.Error("Error running precondition script")
		return ast.WalkStop, err
	}

	if ifResult == ast.WalkSkipChildren {
		theSpinner.Skip("Not required")
		return ast.WalkContinue, err
	}

	r.lastRendered = node

	/***** ENVIRONMENT CAPTURE *****/
	if executionBlock.CaptureEnvironment != nil {
		exec.AddEnvironmentCapture(executionBlock.Language, script, executionBlock.CaptureEnvironment)
	}

	/***** RUN COMMAND *****/
	process, err := runner.Prepare()
	if err != nil {
		return ast.WalkStop, err
	}

	/***** OUTPUT HANDLING *****/
	outputWaiter := sync.WaitGroup{}
	outputStream := process.Stdout
	stderrBuffer := bytes.Buffer{}

	// The output buffer is for showing the STDOUT/STDERR results on error.
	outputBuffer := NewStdoutBuffer()
	outputTargets := []io.Writer{outputBuffer}

	stdoutDisplayTarget := io.Discard

	if executionBlock.ShowStdout {
		stdoutDisplayTarget = indent.NewWriterPipe(w, 4, nil)
	}

	// Setup the screen writer. It also controls RPC functions as some of them affect output, such as spinners.
	screenWriter := NewAnsiScreenWriter(stdoutDisplayTarget)
	outputTargets = append(outputTargets, screenWriter)

	if executionBlock.ShowStdout {
		screenWriter.BeforeFlush(func() { theSpinner.Stop(); theSpinner.StampShadow() })
		screenWriter.AfterFlush(theSpinner.Start)
	}

	screenWriter.CommandHandler = HandleCommands(theSpinner, r.Context)

	// With the output handlers setup, spin up a multiwriter to write to them.
	outputWriters := io.MultiWriter(outputTargets...)
	outputWaiter.Add(1)
	go func() {
		io.Copy(outputWriters, outputStream)
		outputWaiter.Done()
	}()

	// Capture stderr into a buffer too
	outputWaiter.Add(1)
	go func() {
		io.Copy(&stderrBuffer, process.Stderr)
		outputWaiter.Done()
	}()

	/***** WAIT FOR PROCESS TO COMPLETE *****/
	err = process.Start()
	if err != nil {
		return ast.WalkStop, err
	}

	outputWaiter.Wait() // Wait for the process's STDOUT to close.
	exitCode, _, err := process.Wait()

	if err != nil {
		rdutil.Logger.Debug().Msgf("Execution failed with %#v", err)
		return ast.WalkStop, err
	}

	// Flush any remaining writes.
	type flushable interface{ Flush() error }
	for _, t := range outputTargets {
		if flusher, ok := t.(flushable); ok {
			flusher.Flush()
		}
	}

	/***** ERROR HANDLING *****/

	if executionBlock.SkipOnSuccess {
		if err != nil || exitCode != 0 {
			theSpinner.Error("Continue")

			return ast.WalkContinue, nil
		} else {
			theSpinner.Skip("Skip on Success")

			// FIXME - How to skip to the next heading?
			return ast.WalkContinue, nil
		}
	}

	if executionBlock.SkipOnFailure {
		if err != nil || exitCode != 0 {
			theSpinner.Skip(fmt.Sprintf("Skip on Failure (%d)", exitCode))

			skipTo := rundown_ast.GetNextSection(rundown_ast.GetSectionForNode(executionBlock))
			r.skipUntil = skipTo

			return ast.WalkContinue, nil
		}
	}

	if err != nil {
		rdutil.Logger.Debug().Msgf("Error: %s", err.Error())
		return ast.WalkStop, err
	}

	if exitCode != 0 {
		output := ""

		if !executionBlock.ShowStdout {
			output = outputBuffer.String()
		}

		if string(process.StderrOutput) != "" {
			if output != "" {
				output += "\n\n"
			}

			output += string(process.StderrOutput)
		} else {
			stderrBuf := stderrBuffer.String()
			if stderrBuf != "" {
				if output != "" {
					output += "\n\n"
				}

				output += stderrBuf
			}
		}

		output = strings.TrimSpace(output)

		rdutil.Logger.Debug().Msgf("Exit Code is error: %d", exitCode)
		rdutil.Logger.Debug().Msgf("Output is: %s", output)

		theSpinner.Error("Failed")

		r.exitCode = exitCode

		w.WriteString("\n")

		w.WriteString(Aurora.Red("Script Failed:\n").String())

		resultErr := exec.ParseError(script, output)
		r.writeLinesWithPrefix("  ", string(resultErr.String(Aurora)), w)

		// Find on failure nodes
		failureNodes := rundown_ast.GetOnFailureNodes(node)
		insertAfterNode := node
		for _, f := range failureNodes {
			if f.MatchesError([]byte(output)) {
				newNode := f.ConvertToParagraph()
				node.Parent().InsertAfter(node.Parent(), insertAfterNode, newNode)
				insertAfterNode = newNode
			}
		}

		stopFail := rundown_ast.NewStopFail()
		node.Parent().InsertAfter(node.Parent(), insertAfterNode, stopFail)

		// Allow the FailureNode to handle this.
		return ast.WalkContinue, nil
	}

	if executionBlock.CaptureStdoutInto != "" {
		outputTrimmed := strings.TrimSpace(outputBuffer.String())
		r.Context.AddEnv(executionBlock.CaptureStdoutInto, outputTrimmed)
	}

	theSpinner.Success("")

	return ast.WalkContinue, nil
}

func (r *Renderer) renderRundownInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// var buf bytes.Buffer
	// var w2 = bufio.NewWriter(&buf)

	if !node.HasChildren() {
		return ast.WalkSkipChildren, nil
	}

	if entering {

		// renderer := PrepareMarkdown().Renderer()
		// // TODO - Copy all config and apply it to the renderer.

		// rundown := node.(*RundownInline)

		// if rundown.Modifiers.Flags[Flag("ignore")] == true {
		// 	return ast.WalkSkipChildren, nil
		// }

		// for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// 	renderer.Render(w2, source, child)
		// }

		// if r.Config.RundownHandler != nil {
		// 	_, err := r.Config.RundownHandler.OnRundownNode(node, entering)
		// 	if err != nil {
		// 		return ast.WalkStop, err
		// 	}

		// 	s := string(buf.Bytes())
		// 	result, err := r.Config.RundownHandler.Mutate([]byte(s), node)
		// 	if err != nil {
		// 		return ast.WalkStop, err
		// 	}
		// 	w.Write(result)
		// } else {
		// 	w.Write(buf.Bytes())
		// }
	} else {
		if r.Config.RundownHandler != nil {
			_, err := r.Config.RundownHandler.OnRundownNode(node, entering)
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

// Wraps all block outputs.
func (r *Renderer) wrapBlock(renderFunc renderer.NodeRendererFunc) renderer.NodeRendererFunc {
	return func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && r.wrappingWriter == nil {
			r.nonWrappingWriter = writer
			r.wrappingWriter = wordwrap.NewWriter(120)
		}

		bufWriter := bufio.NewWriter(r.wrappingWriter)

		status, err := renderFunc(bufWriter, source, n, entering)

		if !entering {
			r.nonWrappingWriter.WriteString(r.wrappingWriter.String())
			r.nonWrappingWriter.Flush()
			r.wrappingWriter = nil
			r.nonWrappingWriter = nil
		}

		bufWriter.Flush()

		return status, err
	}
}

func (r *Renderer) wrapInline(renderFunc renderer.NodeRendererFunc) renderer.NodeRendererFunc {
	return func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if r.wrappingWriter != nil {
			bufWriter := bufio.NewWriter(r.wrappingWriter)
			status, err := renderFunc(bufWriter, source, n, entering)
			bufWriter.Flush()

			return status, err
		}

		return renderFunc(writer, source, n, entering)
	}
}

func (r *Renderer) supportSkipping(renderFunc renderer.NodeRendererFunc) renderer.NodeRendererFunc {
	return func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if r.skipUntil != nil {
			if n != r.skipUntil {
				rdutil.Logger.Debug().Msgf("Skipping %T", n)
				return ast.WalkSkipChildren, nil
			} else {
				rdutil.Logger.Debug().Msgf("Reached skip target. Rendering this one %T", n)
				r.skipUntil = nil
			}
		}

		if !entering {
			return renderFunc(writer, source, n, entering)
		}

		result, err := r.checkIfScript(n)

		if result == ast.WalkSkipChildren {
			rdutil.Logger.Debug().Msgf("If script indicated failure %T", n)

			r.skipUntil = n.NextSibling()
			if r.skipUntil == nil {
				r.skipUntil = n.OwnerDocument()
			}

			if far, ok := n.(rundown_ast.FarSkip); ok {
				r.skipUntil = far.GetEndSkipNode(n)
			}

			rdutil.Logger.Debug().Msgf("Skipping until %#v", r.skipUntil)

			return result, err
		}

		return renderFunc(writer, source, n, entering)
	}
}

func (r *Renderer) checkIfScript(node ast.Node) (ast.WalkStatus, error) {
	if container, ok := node.(rundown_ast.Conditional); ok {
		rdutil.Logger.Debug().Msgf("Is a conditional: %T", node)
		if container.HasIfScript() {
			rdutil.Logger.Debug().Msgf("Has conditional script: %s", container.GetIfScript())

			if !container.HasResult() {
				result, err := runIfScript(r.Context, container.GetIfScript())

				if err != nil {
					rdutil.Logger.Error().Msgf("Error running conditional: %s", err.Error())
					return ast.WalkStop, err
				}

				container.SetResult(result)
			}

			if !container.GetResult() {
				return ast.WalkSkipChildren, nil
			}
		}
	}

	return ast.WalkContinue, nil
}

// Skip blocks will render their children and then on exit setup to skip.
// This is the reverse of the default skip flow, which skips before rendering children.
func (r *Renderer) renderSkipBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*rundown_ast.SkipBlock)

	if r.skipUntil != nil {
		if n != r.skipUntil {
			rdutil.Logger.Debug().Msgf("Skipping %T", n)
			return ast.WalkSkipChildren, nil
		} else {
			rdutil.Logger.Debug().Msgf("Reached skip target. Rendering this one %T", n)
			r.skipUntil = nil
		}
	}

	if !entering {
		if n.GetResult() {
			r.skipUntil = n.GetEndSkipNode(n)
			rdutil.Logger.Debug().Msgf("Skipping until %#v", r.skipUntil)
		}

		return ast.WalkContinue, nil
	}

	result, err := r.checkIfScript(n)

	if result != ast.WalkSkipChildren {
		rdutil.Logger.Debug().Msgf("If script indicated success %T, rendering", n)

		return ast.WalkContinue, err
	} else {
		rdutil.Logger.Debug().Msgf("If script indicated failure %T, ignoring skip", n)
		return ast.WalkSkipChildren, nil
	}

}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)

	reader := text.NewNodeReaderFromSource(n, source)
	contents, _ := ioutil.ReadAll(reader)
	if bytes.Equal(contents, []byte{}) {
		return ast.WalkSkipChildren, nil
	}

	if entering {
		if _, execBlock := r.lastRendered.(*rundown_ast.ExecutionBlock); execBlock {
			r.writeString(w, "\n")
		}

		style := r.inlineStyles.Push(Color(aurora.CyanFg | aurora.BoldFm))
		r.writeString(w, style.Begin())

		r.writeString(w, strings.Repeat("#", n.Level))
		r.writeString(w, " ")

		r.writeString(w, style.End())

		r.SetLevel(n.Level)
	} else {
		r.inlineStyles.Pop()
		r.writeString(w, "\n\n")
		r.lastRendered = n
	}
	return ast.WalkContinue, nil
}

// BlockquoteAttributeFilter defines attribute names which blockquote elements can have
var BlockquoteAttributeFilter = GlobalAttributeFilter.Extend(
	[]byte("cite"),
)

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.blockStyles.Push(NewBulletSequence(Aurora.Blue(" >").String(), paddingForLevel(r.currentLevel)))
	} else {
		r.blockStyles.Pop()
		r.lastRendered = n
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) syntaxHighlightNode(w util.BufWriter, language string, source []byte, n ast.Node, subEnv bool) {
	target := r.nodeLinesToString(source, n)

	r.syntaxHighlightBytes(w, language, target, subEnv)
}

func (r *Renderer) syntaxHighlightBytes(w util.BufWriter, language string, source string, subEnv bool) {
	lang := lexers.Get(language)

	if subEnv {
		source = rdutil.SubEnv(r.Context.Env, source)
	}

	var buf bytes.Buffer

	if lang == nil {
		lang = lexers.Analyse(source)
	}

	theme := "rundown"
	styles.Register(chroma.MustNewStyle("rundown",
		chroma.StyleEntries{
			chroma.Text:                rundown_styles.Dark.Text,
			chroma.Error:               rundown_styles.Dark.Error,
			chroma.Comment:             rundown_styles.Dark.Comment,
			chroma.CommentPreproc:      rundown_styles.Dark.CommentPreproc,
			chroma.Keyword:             rundown_styles.Dark.Keyword,
			chroma.KeywordReserved:     rundown_styles.Dark.KeywordReserved,
			chroma.KeywordNamespace:    rundown_styles.Dark.KeywordNamespace,
			chroma.KeywordType:         rundown_styles.Dark.KeywordType,
			chroma.Operator:            rundown_styles.Dark.Operator,
			chroma.Punctuation:         rundown_styles.Dark.Punctuation,
			chroma.Name:                rundown_styles.Dark.Name,
			chroma.NameBuiltin:         rundown_styles.Dark.NameBuiltin,
			chroma.NameTag:             rundown_styles.Dark.NameTag,
			chroma.NameAttribute:       rundown_styles.Dark.NameAttribute,
			chroma.NameClass:           rundown_styles.Dark.NameClass,
			chroma.NameConstant:        rundown_styles.Dark.NameConstant,
			chroma.NameDecorator:       rundown_styles.Dark.NameDecorator,
			chroma.NameException:       rundown_styles.Dark.NameException,
			chroma.NameFunction:        rundown_styles.Dark.NameFunction,
			chroma.NameOther:           rundown_styles.Dark.NameOther,
			chroma.Literal:             rundown_styles.Dark.Literal,
			chroma.LiteralNumber:       rundown_styles.Dark.LiteralNumber,
			chroma.LiteralString:       rundown_styles.Dark.LiteralString,
			chroma.LiteralStringEscape: rundown_styles.Dark.LiteralStringEscape,
			chroma.GenericDeleted:      rundown_styles.Dark.GenericDeleted,
			chroma.GenericEmph:         rundown_styles.Dark.GenericEmph,
			chroma.GenericInserted:     rundown_styles.Dark.GenericInserted,
			chroma.GenericStrong:       rundown_styles.Dark.GenericStrong,
			chroma.GenericSubheading:   rundown_styles.Dark.GenericSubheading,
		}))

	if lang != nil {
		formatter := "terminal256"

		if !ColorsEnabled {
			formatter = "noop"
		}

		quick.Highlight(&buf, source, language, formatter, theme)
	} else {
		buf.WriteString(source)
	}

	// Trim any trailing formatting only lines.
	trailing := []string{}
	lines := strings.SplitAfter(buf.String(), "\n")
	for i := len(lines) - 1; i >= 0 && strings.TrimSpace(rdutil.RemoveColors(lines[i])) == ""; {
		trailing = append(trailing, strings.TrimSpace(lines[i]))
		lines = lines[0:i]
		i = i - 1
	}

	r.levelLinesWithPrefix("", strings.TrimSpace(strings.Join(lines, "")), w)
	for i := len(trailing) - 1; i >= 0; i = i - 1 {
		w.WriteString(trailing[i])
	}
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.writeString(w, "\n\n")
		r.levelLinesWithPrefix("", strings.TrimSpace(r.nodeLinesToString(source, n)), w)
	} else {
		_, _ = w.WriteString("\r\n") // Block level element, add blank line.
		r.lastRendered = n
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) ensureBlockSeparator(w util.BufWriter, node ast.Node) {
	render := true

	switch node.PreviousSibling().(type) {
	case *ast.Paragraph:
		// Paragraphs always end with new lines
		render = false
	case nil:
		switch node.Parent().(type) {
		case *rundown_ast.SubEnvBlock:
			render = false // The next call will take care of it.
			r.ensureBlockSeparator(w, node.Parent())
		}
	}

	if render {
		r.writeString(w, "\n\n")
	}
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)

	// If we're embedded inside a SubEnv block, then we want to substitute environment variables.
	_, replacingEnv := node.Parent().(*rundown_ast.SubEnvBlock)

	if entering {
		language := n.Language(source)
		if language != nil {
			r.ensureBlockSeparator(w, node)
			r.syntaxHighlightNode(w, string(language), source, node, replacingEnv)
		}
	} else {
		_, _ = w.WriteString("\r\n") // Block level element, add blank line.
		r.lastRendered = n
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
	return Aurora.Bold(s.marker + " ").String()
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
		r.lastRendered = n
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
		// If we're following an execution block, add a extra line.
		switch r.lastRendered.(type) {
		case *rundown_ast.ExecutionBlock:
			w.WriteString("\n")
		}

		link, ok := n.FirstChild().(*ast.Link)

		if ok && n.ChildCount() == 1 && link.Title == nil {
			// Modifier
			return ast.WalkSkipChildren, nil
		}

		// Indent only if we're inside a list, but not the first child.
		if _, ok := n.Parent().(*ast.ListItem); ok && n.Parent().FirstChild() != n {
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

		r.lastRendered = n
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
		r.lastRendered = n
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

	_, _ = w.WriteString(fmt.Sprintf("  %s  \r\n\r\n", Aurora.Faint(line).String()))
	r.lastRendered = n

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
		r.inlineStyles.Push(Color(aurora.YellowFg))
	} else {
		r.inlineStyles.Pop()
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

func (r *Renderer) renderEnvironmentSubstitution(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		envSub := n.(*rundown_ast.EnvironmentSubstitution)
		name := strings.TrimLeft(string(envSub.Value), "$")

		style := r.inlineStyles.Peek()
		if style != nil {
			r.writeString(w, style.Begin())
		}

		if val, ok := r.Context.Env[name]; ok {

			w.WriteString(val)
		} else {
			w.WriteString(fmt.Sprintf("\nUnknown: %s from %#v\n", name, r.Context.Env))
		}

		if style != nil {
			r.writeString(w, style.End())
		}

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
		// 	_, _ = w.WriteString(Aurora.Faint(" (" + string(n.Title) + ")").String())
		// } else {
		// 	_, _ = w.WriteString(Aurora.Faint(" (").String() + Aurora.Faint(string(n.Destination)).String() + Aurora.Faint(")").String())
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
			if line != "" {
				w.WriteString(line + "\n")
			}

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
	n := node.(*ast.Text)
	if !entering {
		if n.SoftLineBreak() {
			r.writeString(w, " ") // Add a space.
		}

		return ast.WalkContinue, nil
	}

	segment := n.Segment

	if style := r.inlineStyles.Peek(); style != nil {
		r.writeString(w, style.Begin())
	}

	value := segment.Value(source)

	// If the next node is a rundown marker indicating a label, trim
	// any trailing space to present the label properly.
	if _, ok := node.NextSibling().(*rundown_ast.SectionPointer); ok {
		value = bytes.TrimRight(value, " ")
	}

	if n.IsRaw() {
		_, _ = w.Write(value)
	} else {
		_, _ = w.Write(value)
		if n.HardLineBreak() {
			_, _ = w.WriteString("\n")
			r.writeString(w, paddingForLevel(r.currentLevel))
		} else {
			// Do nothing, let word wrapping handle this case.
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
