package ast

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	goldast "github.com/yuin/goldmark/ast"
	"golang.org/x/exp/maps"
	"gopkg.in/guregu/null.v4"
)

type OptionType interface {
	Validate(input string) error       // Validates the normalised input
	Normalise(input string) string     // Treats the input into a standard format
	ResolvedValue(input string) string // Converts input into it's final value to be set into the context
	Describe() string                  // Describes the input and what it accepts
	InputType() string                 // Human relatable input type (string, int, boolean, etc)
}

type OptionTypeRuntime interface {
	NormaliseToPath(input string, path string) (string, error)
}

type TypeBoolean struct{}

type TypeEnum struct {
	ValidValues []string
}

type TypeKV struct {
	Pairs map[string]string
}

type TypeString struct{}

type TypeInt struct{}

type TypeFilename struct {
	MustExist    bool
	MustNotExist bool
}

type TypePath struct{}

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

func BuildOptionType(optionType string) (OptionType, error) {
	optType := strings.ToLower(optionType)

	if strings.HasPrefix(optType, "enum:") {
		options := strings.Split(optionType[5:], "|")
		return &TypeEnum{
			ValidValues: options,
		}, nil
	}

	if strings.HasPrefix(optType, "kv:") {
		pairs := map[string]string{}

		for _, option := range strings.Split(optionType[3:], "|") {
			kv := strings.SplitN(option, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("key-value option `%s` must be in format key=value", option)
			}
			pairs[kv[0]] = kv[1]
		}

		return &TypeKV{
			Pairs: pairs,
		}, nil
	}

	if strings.HasPrefix(optType, "string") {
		return &TypeString{}, nil
	}

	if strings.HasPrefix(optType, "int") || strings.HasPrefix(optType, "number") {
		return &TypeInt{}, nil
	}

	if strings.HasPrefix(optType, "bool") {
		return &TypeBoolean{}, nil
	}

	if strings.HasPrefix(optType, "path") {
		return &TypePath{}, nil
	}

	if strings.HasPrefix(optType, "file:") {
		fileOp := strings.Replace(optType, "file:", "", 1)

		return &TypeFilename{MustExist: fileOp == "exist", MustNotExist: fileOp == "not-exist"}, nil
	} else if optType == "file" {
		return &TypeFilename{}, nil
	}

	return nil, fmt.Errorf("unknown option type `%s`", optType)
}

func (t *TypeKV) InputType() string       { return "string" }
func (t *TypeEnum) InputType() string     { return "string" }
func (t *TypeString) InputType() string   { return "string" }
func (t *TypeBoolean) InputType() string  { return "bool" }
func (t *TypeInt) InputType() string      { return "int" }
func (t *TypeFilename) InputType() string { return "string" }
func (t *TypePath) InputType() string     { return "string" }

func (t *TypeKV) ResolvedValue(input string) string       { return t.Pairs[input] }
func (t *TypeEnum) ResolvedValue(input string) string     { return input }
func (t *TypeString) ResolvedValue(input string) string   { return input }
func (t *TypeBoolean) ResolvedValue(input string) string  { return input }
func (t *TypeInt) ResolvedValue(input string) string      { return input }
func (t *TypeFilename) ResolvedValue(input string) string { return input }
func (t *TypePath) ResolvedValue(input string) string     { return input }

func (t *TypeKV) Normalise(input string) string     { return input }
func (t *TypeEnum) Normalise(input string) string   { return input }
func (t *TypeString) Normalise(input string) string { return input }
func (t *TypeInt) Normalise(input string) string    { return input }
func (t *TypePath) Normalise(input string) string   { return input }

func (t *TypeBoolean) Normalise(input string) string {
	if strings.ToLower(input) == "true" {
		return "true"
	} else {
		return "false"
	}
}

func (t *TypeEnum) Validate(input string) error {
	for _, x := range t.ValidValues {
		if input == x {
			return nil
		}
	}

	return fmt.Errorf("\"%s\" must be one of: %s", input, strings.Join(t.ValidValues, ", "))
}

func (t *TypeKV) Validate(input string) error {
	for k := range t.Pairs {
		if input == k {
			return nil
		}
	}

	return fmt.Errorf("\"%s\" must be one of: %s", input, strings.Join(maps.Keys(t.Pairs), ", "))
}

func (t *TypeString) Validate(string) error {
	return nil
}

func (t *TypeBoolean) Validate(string) error {
	return nil
}

func (t *TypeInt) Validate(input string) error {
	_, err := strconv.Atoi(input)
	return err
}

func (t *TypeEnum) Describe() string {
	return strings.Join(t.ValidValues, ", ")
}

func (t *TypeKV) Describe() string {
	return "one of: " + strings.Join(maps.Keys(t.Pairs), ", ")
}

func (t *TypeString) Describe() string {
	return "any value"
}

func (t *TypeFilename) Describe() string {
	return "any file name"
}

func (t *TypePath) Describe() string {
	return "any path"
}

func (t *TypeBoolean) Describe() string {
	return "true or false"
}

func (t *TypeInt) Describe() string {
	return "a number"
}

func (t *TypeFilename) Validate(string) error {
	return nil
}

func (t *TypePath) Validate(string) error {
	return nil
}

func (t *TypeFilename) Normalise(input string) string {
	return input
}

// Takes the path provided in the option, and treats it as relative to the pwd, returning the absolute path.
func (t *TypeFilename) NormaliseToPath(input string, pwd string) (string, error) {
	rel := path.Join(pwd, input)
	return filepath.Abs(rel)
}

func (t *TypePath) NormaliseToPath(input string, pwd string) (string, error) {
	rel := path.Join(pwd, input)
	return filepath.Abs(rel)
}
