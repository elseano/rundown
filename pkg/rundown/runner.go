package rundown

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var InvocationError = errors.New("InvocationError")

type InvalidOptionsError struct {
	OptionName string
	ShortCode  string
	Message    string
}

func (e *InvalidOptionsError) Error() string {
	if e.ShortCode != "" {
		return fmt.Sprintf("Option '%s+%s' %s", e.ShortCode, e.OptionName, e.Message)
	} else {
		return fmt.Sprintf("'+%s' %s", e.OptionName, e.Message)
	}
}

func (e *InvalidOptionsError) Is(err error) bool {
	return err == InvocationError
}

type InvalidShortCodeError struct {
	ShortCode string
}

func (e *InvalidShortCodeError) Error() string {
	return fmt.Sprintf("Document doesn't have shortcode %s", e.ShortCode)
}

func (e *InvalidShortCodeError) Is(err error) bool {
	return err == InvocationError
}

type DocumentShortCodes struct {
	// The shortcodes defined in the document
	Codes map[string]*ShortCodeInfo

	// The order of the shortcodes for presentation purposes.
	Order []string

	// Functions defined within the document.
	Functions map[string]*ShortCodeInfo

	// Document-wide options defined on the Root section.
	Options map[string]*ShortCodeOption
}

type ShortCodeOption struct {
	Code         string
	Type         string
	Default      string
	Required     bool
	Prompt       bool
	Description  string
	IsPositional bool
	Position     int
}

type ShortCodeInfo struct {
	Code         string
	FunctionName string
	Name         string
	Description  string
	Options      map[string]*ShortCodeOption
	Section      *markdown.Section
}

func (d *DocumentShortCodes) String() string {
	s := ""

	s += fmt.Sprintf("Document:\n")

	for _, opt := range d.Options {
		s += fmt.Sprintf("  %s %s default: %s\n", opt.Code, opt.Type, opt.Default)
	}

	for _, sec := range d.Codes {
		s += fmt.Sprintf("Section %s:\n", sec.Code)

		for _, opt := range sec.Options {
			s += fmt.Sprintf("  %s %s default: %s\n", opt.Code, opt.Type, opt.Default)
		}
	}

	return s
}

type DocumentSpec struct {
	ShortCodes []*ShortCodeSpec
	Options    map[string]*ShortCodeOptionSpec
}

func (d *DocumentSpec) String() string {
	s := ""

	s += fmt.Sprintf("Document:\n")

	for _, opt := range d.Options {
		s += fmt.Sprintf("  %s: %s\n", opt.Code, opt.Value)
	}

	for _, sec := range d.ShortCodes {
		s += fmt.Sprintf("Section %s:\n", sec.Code)

		for _, opt := range sec.Options {
			s += fmt.Sprintf("  %s: %s\n", opt.Code, opt.Value)
		}
	}

	return s
}

type ShortCodeSpec struct {
	Code    string
	Options map[string]*ShortCodeOptionSpec
}

type ShortCodeOptionSpec struct {
	Code  string
	Value string
}

func BuildShortCodeInfo(section *markdown.Section, source []byte) *ShortCodeInfo {
	if section.Label == nil && section.FunctionName == nil {
		return nil
	}

	var (
		shortCodeDescription = ""
		labelStr             = ""
		functionNameStr      = ""
	)

	if section.Label != nil {
		labelStr = *section.Label
	}

	if section.FunctionName != nil {
		functionNameStr = *section.FunctionName
	}

	for descE := section.Description.Front(); descE != nil; descE = descE.Next() {
		desc := descE.Value.(markdown.RundownNode)
		if descriptionValue := desc.GetModifiers().Values[markdown.Parameter("desc")]; descriptionValue != "" {
			shortCodeDescription += descriptionValue
		} else {
			descriptionInner := util.NodeLines(desc, source)
			shortCodeDescription += descriptionInner
		}
	}

	options := BuildOptions(section)

	return &ShortCodeInfo{
		Code:         strings.TrimSpace(labelStr),
		FunctionName: strings.TrimSpace(functionNameStr),
		Name:         strings.TrimSpace(section.Name),
		Description:  strings.TrimSpace(shortCodeDescription),
		Options:      options,
		Section:      section,
	}
}

func BuildOptions(section *markdown.Section) map[string]*ShortCodeOption {
	options := map[string]*ShortCodeOption{}

	for opt := section.Options.FirstChild(); opt != nil; opt = opt.NextSibling() {
		rdOpt := opt.(markdown.RundownNode)

		option := BuildOptionInfo(rdOpt)

		options[option.Code] = option
	}

	return options
}

func BuildOptionInfo(rdOpt markdown.RundownNode) *ShortCodeOption {
	mods := rdOpt.GetModifiers()

	if !mods.HasAny("opt") {
		return nil
	}

	code := strings.TrimSpace(mods.Values[markdown.Parameter("opt")])
	pos := ""
	isPos := false
	posInt := 0

	if (code[0] >= '0' && code[0] <= '9') || code[0] == '*' {
		sp := strings.SplitN(code, ":", 2)
		code = sp[1]
		pos = sp[0]
	}

	if pos == "*" {
		isPos = true
		posInt = -1
	} else if pos != "" {
		isPos = true
		posInt, _ = strconv.Atoi(pos)
	}

	option := &ShortCodeOption{
		Code:         code,
		Type:         strings.TrimSpace(mods.Values[markdown.Parameter("type")]),
		Required:     mods.Flags[markdown.Flag("required")],
		Prompt:       mods.Values[markdown.Parameter("prompt")] != "",
		Description:  strings.TrimSpace(mods.Values[markdown.Parameter("desc")]),
		Default:      strings.TrimSpace(mods.Values[markdown.Parameter("default")]),
		IsPositional: isPos,
		Position:     posInt,
	}

	return option
}

type Runner struct {
	filename     string
	logger       *log.Logger
	out          io.Writer
	consoleWidth int
}

func FromSource(source string) (*Runner, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	f.WriteString(source)
	return &Runner{filename: f.Name(), out: os.Stdout, consoleWidth: -1}, nil
}

func LoadFile(filename string) (*Runner, error) {
	_, err := os.Stat(filename)

	if err != nil {
		return nil, err
	}

	return &Runner{filename: filename, out: os.Stdout, consoleWidth: -1}, nil
}

func (r *Runner) Filename() string {
	return r.filename
}

func (r *Runner) SetOutput(out io.Writer) {
	r.out = out
}

func (r *Runner) SetConsoleWidth(width int) {
	r.consoleWidth = width
}

func (r *Runner) SetLogger(isLogging bool) {
	var debug io.Writer

	if isLogging {
		debug, _ = os.Create("debug.log")
	} else {
		debug = ioutil.Discard
	}

	r.logger = log.New(debug, "", log.Ltime)
}

func (r *Runner) getEngine() (goldmark.Markdown, *Context) {
	ctx := NewContext()
	ctx.Logger = r.logger
	ctx.CurrentFile = r.filename
	if r.consoleWidth > 0 {
		ctx.ConsoleWidth = r.consoleWidth
	}
	ctx.RawOut = r.out

	md, _ := NewRenderer(ctx)

	return md, ctx
}

func (r *Runner) getByteData(filename string) ([]byte, error) {
	// Loads the file, and injects all the import sites.
	finder := regexp.MustCompile(`<r\s+import=["'](.*)["']\s*/>`)
	byteData, err := getByteDataRelative(filename, filename)

	if err != nil {
		util.Debugf("Error loading root file %#v\n", err)
		return nil, err
	}

	for {

		if matches := finder.FindSubmatchIndex(byteData); matches != nil {
			importFilename := string(byteData[matches[2]:matches[3]])
			localData, err := getByteDataRelative(filename, importFilename)

			if err != nil {
				util.Debugf("Error loading imported file %#v\n", err)
				return nil, err
			}

			byteData = bytes.Replace(byteData, byteData[matches[0]:matches[1]], localData, 1)
		} else {
			break
		}
	}

	return byteData, nil
}

func getByteDataRelative(rootFilename string, filename string) ([]byte, error) {
	if filepath.IsAbs(filename) {
		util.Debugf("Loading %s\n", filename)
		return ioutil.ReadFile(filename)
	}

	abs, err := filepath.Abs(rootFilename)
	if err != nil {
		return nil, err
	}

	actualFilename := filepath.Join(filepath.Dir(abs), filename)

	util.Debugf("Loading %s\n", actualFilename)

	return ioutil.ReadFile(actualFilename)
}

func (r *Runner) getAST(md goldmark.Markdown) (ast.Node, []byte) {
	byteData, _ := r.getByteData(r.filename)

	// Trim shebang
	if bytes.HasPrefix(byteData, []byte("#!")) {
		b2 := bytes.SplitN(byteData, []byte("\n"), 2)
		if len(b2) == 2 {
			byteData = b2[1]
		}
	}

	reader := text.NewReader(byteData)
	doc := md.Parser().Parse(reader)

	for d := doc; d != nil; d = d.Parent() {
		doc = d
	}

	return doc, byteData
}

func (r *Runner) GetAST() (ast.Node, []byte) {
	md, _ := r.getEngine()
	doc, source := r.getAST(md)
	return doc, source
}

func (d *DocumentShortCodes) Append(info *ShortCodeInfo) {
	if info.Code != "" {
		d.Codes[info.Code] = info
		d.Order = append(d.Order, info.Code)
	}

	if info.FunctionName != "" {
		d.Functions[info.FunctionName] = info
	}
}

func NewDocumentShortCodes() *DocumentShortCodes {
	return &DocumentShortCodes{
		Options:   map[string]*ShortCodeOption{},
		Codes:     map[string]*ShortCodeInfo{},
		Order:     []string{},
		Functions: map[string]*ShortCodeInfo{},
	}
}

func (r *Runner) GetShortCodes() *DocumentShortCodes {
	md, _ := r.getEngine()
	doc, bytes := r.getAST(md)

	return r.getShortCodes(doc, bytes)
}

func (r *Runner) getShortCodes(doc ast.Node, bytes []byte) *DocumentShortCodes {
	codes := NewDocumentShortCodes()

	if toc, ok := doc.(*markdown.SectionedDocument); ok {
		for _, section := range toc.Sections {
			if info := BuildShortCodeInfo(section, bytes); info != nil {
				codes.Append(info)
			} else if section.Name == "Root" {
				opts := BuildOptions(section)
				codes.Options = opts
			}
		}
	}

	return codes
}

func ParseShortCodeSpecs(specs []string) (*DocumentSpec, error) {
	var (
		result = &DocumentSpec{
			ShortCodes: []*ShortCodeSpec{},
			Options:    map[string]*ShortCodeOptionSpec{},
		}
		currentCode *ShortCodeSpec = nil
	)

	for _, spec := range specs {
		if parts := strings.SplitN(spec, "=", 2); strings.HasPrefix(spec, "--") {
			if len(parts) == 1 {
				return nil, &InvalidOptionsError{OptionName: parts[0], Message: "is required"}
			}

			opt := &ShortCodeOptionSpec{
				Code:  parts[0][1:],
				Value: parts[1],
			}

			if currentCode == nil {
				result.Options[opt.Code] = opt
			} else {
				currentCode.Options[opt.Code] = opt
			}
		} else {
			currentCode = &ShortCodeSpec{
				Code:    spec,
				Options: map[string]*ShortCodeOptionSpec{},
			}

			result.ShortCodes = append(result.ShortCodes, currentCode)
		}
	}

	return result, nil
}

func (r *Runner) RunCodesWithoutValidation(docSpec *DocumentSpec) (error, func()) {
	md, ctx := r.getEngine()
	doc, bytes := r.getAST(md)
	shortCodes := r.getShortCodes(doc, bytes)
	alreadyRun := map[string]bool{}

	for _, opt := range shortCodes.Options {
		optSpec, isSet := docSpec.Options[opt.Code]

		if opt.Default != "" && !isSet {
			optSpec = &ShortCodeOptionSpec{Code: opt.Code, Value: opt.Default}
		}

		if optSpec != nil {
			ctx.SetEnv("OPT_"+strings.ToUpper(opt.Code), optSpec.Value)
		}
	}

	if len(docSpec.ShortCodes) > 0 {

		for _, code := range docSpec.ShortCodes {
			codeDef := shortCodes.Codes[code.Code]
			if codeDef == nil {
				return &InvalidShortCodeError{ShortCode: code.Code}, nil
			}

			section := codeDef.Section
			section.ForceRootLevel()

			for _, opt := range code.Options {
				ctx.SetEnv("OPT_"+strings.ToUpper(opt.Code), opt.Value)
			}

			parents := []*markdown.Section{}
			for n := section.Parent(); n != nil; n = n.Parent() {
				if s, ok := n.(*markdown.Section); ok {
					parents = append(parents, s)
				}
			}

			// Run parent section setups in reverse order, as we collected them by walking up the tree.
			for index := len(parents) - 1; index >= 0; index-- {
				for setup := parents[index].Setups.Front(); setup != nil; setup = setup.Next() {
					setupE := setup.Value.(*markdown.ExecutionBlock)

					if _, ok := alreadyRun[setupE.ID]; !ok {
						err := md.Renderer().Render(ctx.RawOut, bytes, setupE)
						alreadyRun[setupE.ID] = true

						if err != nil {
							return err, nil
						}
					}

				}
			}

			err := md.Renderer().Render(ctx.RawOut, bytes, section)

			if err == nil {
				err = ctx.CurrentError
			}

			if err != nil {
				if stopError, ok := err.(*StopError); ok && stopError.StopHandlers != nil && stopError.StopHandlers.ChildCount() > 0 {
					// Add some space between last node and the output.
					ctx.RawOut.Write([]byte("\r\n"))

					ctx.SetError(stopError)
					return err, func() { md.Renderer().Render(ctx.RawOut, bytes, stopError.StopHandlers) }
				}

				return err, nil
			}

			for _, opt := range code.Options {
				ctx.RemoveEnv("OPT_" + strings.ToUpper(opt.Code))
			}
		}
	} else {
		// w := util.NewCleanNewlineWriter(ctx.RawOut)
		err := md.Renderer().Render(ctx.RawOut, bytes, doc)

		if err == nil {
			err = ctx.CurrentError
		}

		if err != nil {
			if stopError, ok := err.(*StopError); ok && stopError.StopHandlers != nil && stopError.StopHandlers.ChildCount() > 0 {
				// Add some space between last node and the output.
				ctx.RawOut.Write([]byte("\r\n"))

				ctx.SetError(stopError)
				return err, func() { md.Renderer().Render(ctx.RawOut, bytes, stopError.StopHandlers) }
			}

			return err, nil
		}
	}

	return nil, nil
}

func (r *Runner) RunCodes(docSpec *DocumentSpec) (error, func()) {
	md, _ := r.getEngine()
	doc, bytes := r.getAST(md)
	shortCodes := r.getShortCodes(doc, bytes)

	err := ValidateOptions(docSpec, shortCodes)

	if err != nil {
		return err, nil
	}

	return r.RunCodesWithoutValidation(docSpec)

}
