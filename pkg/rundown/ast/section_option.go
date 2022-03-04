package ast

import (
	"fmt"
	"strings"

	goldast "github.com/yuin/goldmark/ast"
	"gopkg.in/guregu/null.v4"
)

type OptionType interface {
	Validate(input string) error
	Normalise(input string) string
	Describe() string
}

type TypeBoolean struct{}

type TypeEnum struct {
	ValidValues []string
}

type TypeString struct{}

type TypeFilename struct {
	MustExist    bool
	MustNotExist bool
}

type SectionOption struct {
	goldast.BaseInline
	OptionName        string
	OptionType        OptionType
	OptionTypeString  string
	OptionDescription string
	OptionPrompt      null.String
	OptionDefault     null.String
	OptionRequired    bool
	OptionAs          string
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionOption(name string) *SectionOption {
	return &SectionOption{
		OptionName: name,
		OptionAs:   toEnvName(name),
	}
}

func toEnvName(name string) string {
	return "OPT_" + strings.ReplaceAll(strings.ToUpper(name), "-", "_")
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionOption = goldast.NewNodeKind("SectionOption")

// Kind implements Node.Kind.
func (n *SectionOption) Kind() goldast.NodeKind {
	return KindSectionOption
}

func (n *SectionOption) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{
		"OptionName":  n.OptionName,
		"Type":        fmt.Sprintf("%#v", n.OptionType),
		"Required":    boolToStr(n.OptionRequired),
		"Prompt":      n.OptionPrompt.ValueOrZero(),
		"WillPrompt":  boolToStr(n.OptionPrompt.Valid),
		"Default":     n.OptionDefault.ValueOrZero(),
		"Description": n.OptionDescription,
	}, nil)
}

func BuildOptionType(optionType string) OptionType {
	optType := strings.ToLower(optionType)

	if strings.HasPrefix(optType, "enum|") {
		options := strings.Split(optType, "|")
		return &TypeEnum{
			ValidValues: options[1:],
		}
	}

	if strings.HasPrefix(optType, "string") {
		return &TypeString{}
	}

	if strings.HasPrefix(optType, "bool") {
		return &TypeBoolean{}
	}

	if strings.HasPrefix(optType, "file-exists") {
		return &TypeFilename{MustExist: true}
	}

	return nil
}

func (t *TypeEnum) Validate(input string) error {
	for _, x := range t.ValidValues {
		if input == x {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", strings.Join(t.ValidValues, ", "))
}

func (t *TypeEnum) Normalise(string) string {
	return ""
}

func (t *TypeString) Validate(string) error {
	return nil
}

func (t *TypeString) Normalise(string) string {
	return ""
}

func (t *TypeBoolean) Validate(string) error {
	return nil
}

func (t *TypeEnum) Describe() string {
	return strings.Join(t.ValidValues, ", ")
}

func (t *TypeString) Describe() string {
	return "any value"
}

func (t *TypeFilename) Describe() string {
	return "any file name"
}

func (t *TypeBoolean) Describe() string {
	return "true or false"
}

func (t *TypeBoolean) Normalise(input string) string {
	if strings.ToLower(input) == "true" {
		return "true"
	} else {
		return "false"
	}
}

func (t *TypeFilename) Validate(string) error {
	return nil
}

func (t *TypeFilename) Normalise(input string) string {
	return input
}
