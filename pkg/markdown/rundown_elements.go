package markdown

import (
	"bytes"
	"container/list"
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

/*
 *
 *   RUNDOWN EXTENSION
 *
 */

type rundownElements struct {
}

var RundownElements = &rundownElements{}

func (e *rundownElements) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewRundownInlineParser(), 2),
			// util.Prioritized(NewTextSubParser(), 2),
		),
		parser.WithBlockParsers(
			util.Prioritized(NewRundownBlockParser(), 2),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewRundownASTTransformer(), 1),
		),
		parser.WithParagraphTransformers(
			util.Prioritized(NewRundownParaTransformer(), 1),
		),
	)
}

type RundownNode interface {
	ast.Node
	GetModifiers() *Modifiers
}

/*
 *
 *   RUNDOWN INLINE NODE
 *
 */

// A Rundown struct represents an inline <rundown> element.
type RundownInline struct {
	ast.BaseInline
	Segments       *text.Segments
	Modifiers      *Modifiers
	MutateContents func(input []byte) []byte
}

func (n *RundownInline) GetModifiers() *Modifiers {
	return n.Modifiers
}

// Inline implements Inline.Inline.
func (n *RundownInline) Inline() {}

// Dump implements Node.Dump.
func (n *RundownInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Modifiers": n.Modifiers.String()}, nil)
}

// KindRundown is a NodeKind of the Rundown node.
var KindRundownInline = ast.NewNodeKind("RundownInline")

// Kind implements Node.Kind.
func (n *RundownInline) Kind() ast.NodeKind {
	return KindRundownInline
}

// NewRundown returns a new Rundown node.
func NewRundownInline(mods *Modifiers) *RundownInline {
	return &RundownInline{
		BaseInline: ast.BaseInline{},
		Segments:   text.NewSegments(),
		Modifiers:  mods,
	}
}

/*
 *
 *   DETECTOR
 *
 */

var attributePattern = regexp.MustCompile(`(?:\s+[a-zA-Z_:][a-zA-Z0-9:._-]*(?:\s*=\s*(?:[^\"'=<>` + "`" + `\x00-\x20]+|'[^']*'|"[^"]*"))?)`)
var rundownTagRegexp = regexp.MustCompile(`^(?:<(r|rundown))\s*(.*?)\s*(?:(/>)|>)\s*`)

var rundownTagClose = regexp.MustCompile(`</r>|</rundown>`)

func matchRundownTag(block text.Reader, pc parser.Context) []int {
	line, _ := block.PeekLine()
	match := rundownTagRegexp.FindSubmatchIndex(line)

	if match == nil || len(match) == 0 {
		return nil
	}

	return []int{match[0], match[1], match[4], match[5], match[6], match[7]}
}

/*
 *
 *   RUNDOWN INLINE PARSER
 *
 */

type RundownInlineParser struct {
}

var defaultRundownInlineParser = &RundownInlineParser{}

// NewRundownInlineParser return a new InlineParser that can parse
// inline htmls
func NewRundownInlineParser() parser.InlineParser {
	return defaultRundownInlineParser
}

func (s *RundownInlineParser) Trigger() []byte {
	return []byte{'<'}
}

var rundownInlineStateKey = parser.NewContextKey()
var rundownBottom = parser.NewContextKey()

type rundownInlineState struct {
	ast.BaseInline
	Modifiers *Modifiers
}

func (s *rundownInlineState) Text(source []byte) []byte {
	return []byte{}
}

func (s *rundownInlineState) Dump(source []byte, level int) {
	fmt.Printf("%srundownInlineState: \"%s\"\n", strings.Repeat("    ", level), s.Text(source))
}

var kindRundownInlineState = ast.NewNodeKind("RundownInlineState")

func (s *rundownInlineState) Kind() ast.NodeKind {
	return kindRundownInlineState
}

func newRundownInlineState() *rundownInlineState {
	return &rundownInlineState{
		Modifiers: NewModifiers(),
	}
}

func (s *RundownInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	if bytes.HasPrefix(line, []byte("<r")) {
		if match := matchRundownTag(block, pc); match != nil {

			mods := ParseModifiers(string(line[match[2]:match[3]]), "=")

			// Moved into the Console Renderer.
			// // Is the parent a Heading? Remove the space between tag and text.
			// if _, ok := parent.(*ast.Heading); ok {
			// 	if subject, ok := parent.LastChild().(*ast.Text); ok {
			// 		cutSeg := subject.Segment.WithStop(segment.Start)
			// 		trimmedSeg := cutSeg.TrimRightSpace(block.Source())
			// 		subject.Segment = trimmedSeg
			// 	}
			// }

			// node.Segments.Append(segment.WithStop(segment.Start + match[1]))

			if match[4] > -1 { // Short closing tag attached
				block.Advance(match[5]) // Move to the end of the closing tag.
				return NewRundownInline(mods)
			}

			// Bookmark where we are so we can proess delimiters later.
			pc.Set(rundownBottom, pc.LastDelimiter())

			node := newRundownInlineState()
			node.Modifiers.Ingest(mods)

			pc.Set(rundownInlineStateKey, node)

			block.Advance(match[1])

			return node
		}
	} else if bytes.HasPrefix(line, []byte("</r")) {
		if match := rundownTagClose.FindIndex(line); match != nil && match[0] == 0 {
			block.Advance(match[1])
			state, ok := pc.Get(rundownInlineStateKey).(*rundownInlineState)
			if !ok {
				return nil
			}

			pc.Set(rundownInlineStateKey, nil)

			node := NewRundownInline(state.Modifiers)

			// Process contents.
			var bottom ast.Node
			if v := pc.Get(rundownBottom); v != nil {
				bottom = v.(ast.Node)
			}
			pc.Set(rundownBottom, nil)
			parser.ProcessDelimiters(bottom, pc)

			// Move contents into the Rundown node.
			for c := state.NextSibling(); c != nil; {
				next := c.NextSibling()
				parent.RemoveChild(parent, c)
				node.AppendChild(node, c)
				c = next
			}

			// Remove the state element
			state.Parent().RemoveChild(state.Parent(), state)

			return node
		}
	}

	return nil
}

func (s *RundownInlineParser) CloseBlock(parent ast.Node, pc parser.Context) {
	// nothing to do
}

/*
 *
 *   RUNDOWN BLOCK NODE
 *
 */

type RundownBlock struct {
	ast.BaseBlock

	ForCodeBlock bool
	Modifiers    *Modifiers
}

func (n *RundownBlock) GetModifiers() *Modifiers {
	return n.Modifiers
}

// IsRaw implements Node.IsRaw.
func (n *RundownBlock) IsRaw() bool {
	return true
}

// Dump implements Node.Dump.
func (n *RundownBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Modifiers": n.Modifiers.String()}, nil)
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindRundownBlock = ast.NewNodeKind("RundownBlock")

// Kind implements Node.Kind.
func (n *RundownBlock) Kind() ast.NodeKind {
	return KindRundownBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewRundownBlock(modifiers *Modifiers) *RundownBlock {
	return &RundownBlock{
		BaseBlock:    ast.BaseBlock{},
		Modifiers:    modifiers,
		ForCodeBlock: false,
	}
}

/*
 *
 *   Text Substitution Inline
 *
 */

type TextSubInline struct {
	ast.BaseInline
	Segment  text.Segment
	contents []byte
}

// Dump implements Node.Dump.
func (n *TextSubInline) Dump(source []byte, level int) {
	fmt.Printf("%sTextSubstitution: \"%s\"\n", strings.Repeat("    ", level), strings.TrimRight(string(n.Text(source)), "\n"))
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindTextSubInline = ast.NewNodeKind("TextSubInline")

// Kind implements Node.Kind.
func (n *TextSubInline) Kind() ast.NodeKind {
	return KindTextSubInline
}

// NewRundownBlock returns a new RundownBlock node.
func NewTextSubInline() *TextSubInline {
	return &TextSubInline{
		BaseInline: ast.BaseInline{},
	}
}

func (n *TextSubInline) Text(source []byte) []byte {
	if n.contents != nil {
		return n.contents
	}

	return n.Segment.Value(source)
}

func (n *TextSubInline) Substitute(contents []byte) {
	n.contents = contents
}

/*
 *
 * Text Sub Inline Parser
 *
 */

type textSubDelimiterProcessor struct {
}

func (p *textSubDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '$' || !bytes.ContainsAny([]byte{b}, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
}

func (p *textSubDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *textSubDelimiterProcessor) OnMatch(consumes int) ast.Node {
	return NewTextSubInline()
}

var defaultTextSubDelimiterProcessor = &textSubDelimiterProcessor{}

type textSubParser struct {
}

var defaultTextSubParser = &textSubParser{}

// NewEmphasisParser return a new InlineParser that parses emphasises.
func NewTextSubParser() parser.InlineParser {
	return defaultTextSubParser
}

func (s *textSubParser) Trigger() []byte {
	return []byte{'$'}
}

func (s *textSubParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	end := 0

	for i := 1; i < len(line); i++ {
		c := line[i]

		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c == '_') {
			end = i + 1
			continue
		}

		break
	}

	if end == 0 {
		return nil
	}

	node := NewTextSubInline()
	node.Segment = segment.WithStop(segment.Start + end)

	block.Advance(end)

	return node
}

/*
 *
 *   RUNDOWN BLOCK PARSER
 *
 */

type rundownBlockParser struct {
}

var defaultRundownBlockParser = &rundownBlockParser{}

func NewRundownBlockParser() parser.BlockParser {
	return defaultRundownBlockParser
}

func (b *rundownBlockParser) Trigger() []byte {
	return []byte{'<'}
}

func (b *rundownBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	var node *RundownBlock
	line, segment := reader.PeekLine()

	if pos := pc.BlockOffset(); pos < 0 || line[pos] != '<' {
		return nil, parser.NoChildren
	}

	// Check to see if block has no content
	noContentBlock := bytes.HasSuffix(line, []byte("/>\n")) ||
		bytes.HasSuffix(line, []byte("></r>")) ||
		bytes.HasSuffix(line, []byte("></rundown>"))

	// If this block has content, then return nil. This will turn it into a paragraph,
	// and let the RundownInlineParser pick it up.
	if !noContentBlock {
		return nil, parser.NoChildren
	}

	if bytes.HasPrefix(line, []byte("<r")) {
		if match := matchRundownTag(reader, pc); match != nil {
			mods := ParseModifiers(string(line[match[2]:match[3]]), "=")
			node = NewRundownBlock(mods)
		}
	}

	if node != nil {
		reader.Advance(segment.Len() - 1)
		node.Lines().Append(segment)
		return node, parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *rundownBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *rundownBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *rundownBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *rundownBlockParser) CanAcceptIndentedLine() bool {
	return false
}

/*
 *
 * EXECUTION BLOCK
 *
 * Represents code to be executed.
 *
 */

type ExecutionBlock struct {
	ast.BaseBlock

	Modifiers *Modifiers
	Syntax    string
	Origin    *ast.FencedCodeBlock
	ID        string
}

// IsRaw implements Node.IsRaw.
func (n *ExecutionBlock) SetOrigin(fcb *ast.FencedCodeBlock, source []byte) {
	n.Origin = fcb

	if fcb.Lines().Len() == 0 {
		n.ID = "0:000"
	} else {
		h := md5.New()
		h.Write(fcb.Text(source))
		m := h.Sum(nil)
		m2 := m[len(m)-5:]

		n.ID = fmt.Sprintf("%d:%x", fcb.Lines().At(0).Start, m2)
	}
}

// Dump implements Node.Dump.
func (n *ExecutionBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Modifiers": n.Modifiers.String(),
		"ID":        n.ID,
		"Syntax":    n.Syntax,
	}, nil)
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindExecutionBlock = ast.NewNodeKind("ExecutionBlock")

// Kind implements Node.Kind.
func (n *ExecutionBlock) Kind() ast.NodeKind {
	return KindExecutionBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewExecutionBlock(syntax string, modifiers *Modifiers) *ExecutionBlock {
	return &ExecutionBlock{
		BaseBlock: ast.BaseBlock{},
		Modifiers: modifiers,
		Syntax:    syntax,
	}
}

/*
 *
 * SECTION CONTAINER
 *
 * Groups heading and contents together.
 *
 */

type Section struct {
	ast.BaseBlock

	Handlers    *Container
	Options     *Container
	Description list.List // Description is a list as we want to keep it inside the DOM.
	Setups      list.List // Setups is a list to keep them in the DOM too.

	lastDoc *ast.Document

	Label        *string
	FunctionName *string
	Level        int
	Name         string
}

// Forces the section to be at Level 1, and all children are shifted accordingly.
func (n *Section) ForceRootLevel() {
	levelDelta := -n.Level + 1

	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node.Kind() {
			case ast.KindHeading:
				node.(*ast.Heading).Level += levelDelta
			case KindSection:
				node.(*Section).Level += levelDelta
			}
		}

		return ast.WalkContinue, nil
	})
}

// Shifts to Root Level, and then adds the given Level to all.
func (n *Section) ForceLevel(newLevel int) {
	n.ForceRootLevel()

	// Because root level is level 1
	levelDelta := newLevel - 1

	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node.Kind() {
			case ast.KindHeading:
				node.(*ast.Heading).Level += levelDelta
			case KindSection:
				node.(*Section).Level += levelDelta
			}
		}

		return ast.WalkContinue, nil
	})
}

// Dump implements Node.Dump.
func (n *Section) Dump(source []byte, level int) {
	label := "(not set)"
	if n.Label != nil {
		label = *n.Label
	}

	functionName := "(not set)"
	if n.FunctionName != nil {
		functionName = *n.FunctionName
	}

	descList := []string{}
	for descE := n.Description.Front(); descE != nil; descE = descE.Next() {
		descList = append(descList, string(descE.Value.(ast.Node).Text(source)))
	}

	ast.DumpHelper(n, source, level, map[string]string{
		"Level":        fmt.Sprintf("%d", n.Level),
		"Label":        fmt.Sprintf("%s", label),
		"FunctionName": fmt.Sprintf("%s", functionName),
		"Description":  fmt.Sprintf("%#v", descList),
		"Name":         n.Name,
	}, func(subLevel int) {
		n.Handlers.Dump(source, subLevel)
		n.Options.Dump(source, subLevel)

		for setup := n.Setups.Front(); setup != nil; setup = setup.Next() {
			fmt.Printf("%sSetups {\n", strings.Repeat("    ", subLevel))
			setup.Value.(ast.Node).Dump(source, subLevel+1)
			fmt.Printf("%s}\n", strings.Repeat("    ", subLevel))
		}
	})
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSection = ast.NewNodeKind("Section")

// Kind implements Node.Kind.
func (n *Section) Kind() ast.NodeKind {
	return KindSection
}

func NewSectionForRoot() *Section {
	section := &Section{
		BaseBlock:    ast.BaseBlock{},
		Label:        nil,
		FunctionName: nil,
		Level:        0,
		Name:         "Root",
		Handlers:     NewContainer("Handlers"),
		Options:      NewContainer("Options"),
	}

	return section
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionFromHeading(heading *ast.Heading, source []byte) *Section {
	// Find rundown child
	var (
		rundown      *RundownInline = nil
		label        *string
		functionName *string
	)

	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if rd, ok := child.(*RundownInline); ok {
			rundown = rd
			break
		}
	}

	if rundown != nil {
		label = rundown.Modifiers.GetValue(Parameter("label"))
	}

	if rundown != nil {
		functionName = rundown.Modifiers.GetValue(Parameter("func"))
	}

	section := &Section{
		BaseBlock:    ast.BaseBlock{},
		Label:        label,
		FunctionName: functionName,
		Level:        heading.Level,
		Name:         string(heading.Text(source)),
		Handlers:     NewContainer("Handlers"),
		Options:      NewContainer("Options"),
	}

	return section
}

func (n *Section) appendHandler(rundown RundownNode) {
	n.Handlers.AppendChild(n.Handlers, rundown)
}

func (n *Section) appendOption(rundown RundownNode) {
	n.Options.AppendChild(n.Options, rundown)

	if rundown.GetModifiers().HasAny("prompt") {
		rundownCopy := NewRundownBlock(rundown.GetModifiers())
		n.AppendChild(n, rundownCopy)
	}
}

func (n *Section) appendDesc(rundown RundownNode) {
	n.Description.PushBack(rundown)
}

func (n *Section) appendSetup(exec *ExecutionBlock) {
	n.Setups.PushBack(exec)

	// Add the Description element, as we also consider it to be part of the
	// normal content flow.
	n.AppendChild(n, exec)
}

func (n *Section) Append(child ast.Node) {
	if rundown, ok := child.(*RundownBlock); ok {
		switch {
		case rundown.Modifiers.HasAny("on-failure"):
			n.appendHandler(rundown)
			return
		case rundown.Modifiers.HasAll("opt", "desc"):
			n.appendOption(rundown)
			return
		case rundown.Modifiers.HasAny("desc"):
			n.appendDesc(rundown)

			// Add the Description element, as we also consider it to be part of the
			// normal content flow.
			n.AppendChild(n, rundown)

			return
		}
	} else if exec, ok := child.(*ExecutionBlock); ok {
		if exec.Modifiers.HasAny("setup") {
			n.appendSetup(exec)
		}
	} else if para, ok := child.(*ast.Paragraph); ok {
		// Search for any embedded description nodes.
		ast.Walk(para, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if rd, ok := node.(*RundownInline); ok {
					if rd.Modifiers.HasAny("desc") {
						// Only need to add to the desc, as the parent block will be added to the tree.
						n.appendDesc(rd)
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}

	n.AppendChild(n, child)
}

func terminatesSectionDoc(node ast.Node) bool {
	switch node.Kind() {
	case KindSection:
		return true
	case KindExecutionBlock:
		return true
	default:
		node, ok := node.(*RundownBlock)

		return ok && (node.Modifiers.HasAll("invoke") || node.Modifiers.HasAll("opt", "prompt"))
	}
}

func (n *Section) AppendChild(self, node ast.Node) {
	if terminatesSectionDoc(node) {
		n.BaseBlock.AppendChild(self, node)
		n.lastDoc = nil
	} else {
		if n.lastDoc == nil {
			n.lastDoc = ast.NewDocument()
			n.BaseBlock.AppendChild(self, n.lastDoc)
		}

		n.lastDoc.AppendChild(n.lastDoc, node)
	}
}

type Container struct {
	ast.Document
	Name string
}

var KindContainer = ast.NewNodeKind("Container")

// Kind implements Node.Kind.
func (n *Container) Kind() ast.NodeKind {
	return KindContainer
}

func NewContainer(name string) *Container {
	return &Container{
		Document: *ast.NewDocument(),
		Name:     name,
	}
}

func (n *Container) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Name": n.Name,
	}, nil)
}

/*
 *
 * SECTIONEDDOCUMENT CONTAINER
 *
 * Maintains an index of sections within the document.
 *
 */

type SectionedDocument struct {
	ast.Document
	Sections []*Section
}

// Dump implements Node.Dump.
func (n *SectionedDocument) Dump(source []byte, level int) {
	functions := []string{}
	sections := []string{}

	for _, s := range n.Sections {
		if s.FunctionName != nil {
			functions = append(functions, *s.FunctionName)
		}
		if s.Label != nil {
			sections = append(sections, *s.Label)
		}
	}

	ast.DumpHelper(n, source, level, map[string]string{
		"Functions":  strings.Join(functions, ", "),
		"ShortCodes": strings.Join(sections, ", "),
	}, nil)
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionedDocument = ast.NewNodeKind("SectionedDocument")

// Kind implements Node.Kind.
func (n *SectionedDocument) Kind() ast.NodeKind {
	return KindSectionedDocument
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionedDocument() *SectionedDocument {
	return &SectionedDocument{
		Document: *ast.NewDocument(),
		Sections: []*Section{},
	}
}

func (n *SectionedDocument) AddSection(section *Section) {
	n.Sections = append(n.Sections, section)
}

/*
 *
 * Rundown AST Transformer
 * - Moves FCB modifiers into dedicated Rundown blocks.
 * - Handles loose RundownBlocks
 * - Builds Section container nodes & rearranges handlers.
 * - Builds ExecutionBlock nodes.
 *
 */

type rundownParagraphTransformer struct {
}

func NewRundownParaTransformer() parser.ParagraphTransformer {
	return &rundownParagraphTransformer{}
}

func (t *rundownParagraphTransformer) Transform(para *ast.Paragraph, reader text.Reader, pc parser.Context) {
	ast.Walk(para, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if rd, ok := node.(*RundownInline); ok {
			if val, ok := rd.Modifiers.Flags[Flag("sub-env")]; val && ok {
				subEnv := NewTextSubInline()

				for child := rd.FirstChild(); child != nil; child = child.NextSibling() {
					subEnv.AppendChild(subEnv, child)
				}

				rd.AppendChild(rd, subEnv)
			}
		}

		return ast.WalkContinue, nil
	})
}
