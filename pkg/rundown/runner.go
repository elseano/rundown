package rundown

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type DocumentShortCodes struct {
	Codes     map[string]*ShortCodeInfo
	Order     []string
	Functions map[string]*ShortCodeInfo
}

type ShortCodeOption struct {
	Code        string
	Type        string
	Default     string
	Required    bool
	Description string
}

type ShortCodeInfo struct {
	Code         string
	FunctionName string
	Name         string
	Description  string
	Options      map[string]*ShortCodeOption
	Section      *markdown.Section
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
		options              = map[string]*ShortCodeOption{}
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

	for opt := section.Options.FirstChild(); opt != nil; opt = opt.NextSibling() {
		rdOpt := opt.(markdown.RundownNode)

		option := &ShortCodeOption{
			Code:        strings.TrimSpace(rdOpt.GetModifiers().Values[markdown.Parameter("opt")]),
			Type:        strings.TrimSpace(rdOpt.GetModifiers().Values[markdown.Parameter("type")]),
			Required:    rdOpt.GetModifiers().Flags[markdown.Flag("required")],
			Description: strings.TrimSpace(rdOpt.GetModifiers().Values[markdown.Parameter("desc")]),
			Default:     strings.TrimSpace(rdOpt.GetModifiers().Values[markdown.Parameter("default")]),
		}

		options[option.Code] = option
	}

	return &ShortCodeInfo{
		Code:         strings.TrimSpace(labelStr),
		FunctionName: strings.TrimSpace(functionNameStr),
		Name:         strings.TrimSpace(section.Name),
		Description:  strings.TrimSpace(shortCodeDescription),
		Options:      options,
		Section:      section,
	}
}

type Runner struct {
	filename string
	logger   *log.Logger
}

func LoadFile(filename string) (*Runner, error) {
	return &Runner{filename: filename}, nil
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

func (r *Runner) getEngine() (goldmark.Markdown, *util.WordWrapWriter, *Context) {
	md := markdown.PrepareMarkdown()
	renderer := md.Renderer()
	ctx := NewContext()
	ctx.Logger = r.logger
	ctx.CurrentFile = r.filename

	currentLevel := 0

	renderer.AddOptions(markdown.WithRundownHandler(NewRundownHandler(ctx)))
	renderer.AddOptions(markdown.WithConsoleWidth(ctx.ConsoleWidth))
	renderer.AddOptions(markdown.WithLevelChange(func(indent int) {
		currentLevel = indent
	}))
	renderer.AddOptions(markdown.WithLevel(currentLevel))

	www := util.NewWordWrapWriter(os.Stdout, ctx.ConsoleWidth)

	www.SetAfterWrap(func(out io.Writer) int {
		n, _ := out.Write(bytes.Repeat([]byte("  "), currentLevel-1))
		return n
	})

	return md, www, ctx
}

func (r *Runner) getAST(md goldmark.Markdown) (ast.Node, []byte) {
	bytes, _ := ioutil.ReadFile(r.filename)
	reader := text.NewReader(bytes)
	doc := md.Parser().Parse(reader)

	for d := doc; d != nil; d = d.Parent() {
		doc = d
	}

	return doc, bytes
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
		Codes:     map[string]*ShortCodeInfo{},
		Order:     []string{},
		Functions: map[string]*ShortCodeInfo{},
	}
}

func (r *Runner) GetShortCodes() *DocumentShortCodes {
	md, _, _ := r.getEngine()
	doc, bytes := r.getAST(md)

	return r.getShortCodes(doc, bytes)
}

func (r *Runner) getShortCodes(doc ast.Node, bytes []byte) *DocumentShortCodes {
	codes := NewDocumentShortCodes()

	if toc, ok := doc.(*markdown.SectionedDocument); ok {
		for _, section := range toc.Sections {
			if info := BuildShortCodeInfo(section, bytes); info != nil {
				codes.Append(info)
			}
		}
	}

	return codes
}

func ParseShortCodeSpecs(specs []string) ([]*ShortCodeSpec, error) {
	var (
		result                     = []*ShortCodeSpec{}
		currentCode *ShortCodeSpec = nil
	)

	for _, spec := range specs {
		if parts := strings.SplitN(spec, "=", 2); strings.HasPrefix(spec, "+") {
			if currentCode == nil {
				return nil, errors.New("Option " + parts[0] + " specified before ShortCode")
			}

			if len(parts) == 1 {
				return nil, errors.New("Option " + parts[0] + " requires value")
			}

			opt := &ShortCodeOptionSpec{
				Code:  parts[0][1:],
				Value: parts[1],
			}

			currentCode.Options[opt.Code] = opt
		} else {
			currentCode = &ShortCodeSpec{
				Code:    spec,
				Options: map[string]*ShortCodeOptionSpec{},
			}

			result = append(result, currentCode)
		}
	}

	return result, nil
}

func (r *Runner) RunCodes(codeArgs []*ShortCodeSpec) error {
	md, www, ctx := r.getEngine()
	doc, bytes := r.getAST(md)
	shortCodes := r.getShortCodes(doc, bytes)

	if len(shortCodes.Codes) == 0 {
		return errors.New("Document does not support ShortCodes")
	}

	for _, code := range codeArgs {
		section := shortCodes.Codes[code.Code]
		if section == nil {
			return errors.New("Invalid ShortCode: " + code.Code)
		}

		for _, opt := range code.Options {
			if section.Options[opt.Code] == nil {
				return errors.New("Invalid ShortCode Option: " + code.Code + " +" + opt.Code)
			}
		}

		for _, opt := range section.Options {
			_, isSet := code.Options[opt.Code]

			if opt.Required && opt.Default == "" && !isSet {
				return errors.New(fmt.Sprintf("ShortCode %s requires option %s to be specified", code.Code, opt.Code))
			}

			if opt.Default != "" && !isSet {
				code.Options[opt.Code] = &ShortCodeOptionSpec{Code: opt.Code, Value: opt.Default}
			}
		}
	}

	for _, code := range codeArgs {
		section := shortCodes.Codes[code.Code].Section
		section.ForceRootLevel()

		for _, opt := range code.Options {
			ctx.SetEnv("OPT_"+strings.ToUpper(opt.Code), opt.Value)
		}

		err := md.Renderer().Render(www, bytes, section)
		if err != nil {
			return err
		}

		for _, opt := range code.Options {
			ctx.RemoveEnv("OPT_" + strings.ToUpper(opt.Code))
		}
	}

	return nil
}

func (r *Runner) RunSequential() error {
	md, www, _ := r.getEngine()
	doc, bytes := r.getAST(md)

	err := md.Renderer().Render(www, bytes, doc)

	if err != nil {
		return err
	}

	return nil
}
