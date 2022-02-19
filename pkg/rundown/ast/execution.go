package ast

import (
	"crypto/rand"
	"fmt"
	"math/big"

	goldast "github.com/yuin/goldmark/ast"
)

type SpinnerMode int

const (
	SpinnerModeVisible SpinnerMode = iota
	SpinnerModeHidden
	SpinnerModeInlineAll
	SpinnerModeInlineFirst
)

const DefaultSpinnerName string = "Running..."

type ExecutionBlock struct {
	goldast.BaseBlock
	CodeBlock *goldast.FencedCodeBlock

	ID                    string
	ShowStdout            bool
	ShowStderr            bool
	Reveal                bool
	Execute               bool
	CaptureEnvironment    bool
	SubstituteEnvironment bool
	SpinnerName           string
	SpinnerMode           SpinnerMode
	With                  string
	ReplaceProcess        bool
}

// NewRundownBlock returns a new RundownBlock node.
func NewExecutionBlock(fcb *goldast.FencedCodeBlock) *ExecutionBlock {
	id, _ := rand.Int(rand.Reader, big.NewInt(100000000))

	return &ExecutionBlock{
		ID:          fmt.Sprintf("%d", id),
		BaseBlock:   goldast.NewParagraph().BaseBlock,
		CodeBlock:   fcb,
		SpinnerMode: SpinnerModeVisible,
		SpinnerName: DefaultSpinnerName,
		Execute:     true,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindExecutionBlock = goldast.NewNodeKind("ExecutionBlock")

// Kind implements Node.Kind.
func (n *ExecutionBlock) Kind() goldast.NodeKind {
	return KindExecutionBlock
}

func (n *ExecutionBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{
		"ShowStdout":            boolToStr(n.ShowStdout),
		"ShowStderr":            boolToStr(n.ShowStderr),
		"Reveal":                boolToStr(n.Reveal),
		"Execute":               boolToStr(n.Execute),
		"CaptureEnvironment":    boolToStr(n.CaptureEnvironment),
		"SubstituteEnvironment": boolToStr(n.SubstituteEnvironment),
		"ReplaceProcess":        boolToStr(n.ReplaceProcess),
		"SpinnerName":           n.SpinnerName,
		"SpinnerMode":           fmt.Sprintf("%#v", n.SpinnerMode),
		"With":                  n.With,
	}, nil)
}

func boolToStr(in bool) string {
	if in {
		return "true"
	} else {
		return "false"
	}
}
