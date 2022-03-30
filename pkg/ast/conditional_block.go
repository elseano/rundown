package ast

import (
	"github.com/elseano/rundown/pkg/util"
	goldast "github.com/yuin/goldmark/ast"
)

type ConditionalStart struct {
	goldast.BaseBlock
	ConditionalImpl
	ID  string
	End *ConditionalEnd
}

type ConditionalEnd struct {
	goldast.BaseBlock
	Start *ConditionalStart
}

// NewRundownBlock returns a new RundownBlock node.
func NewConditionalStart() *ConditionalStart {
	start := &ConditionalStart{
		ID:  util.RandomString(),
		End: &ConditionalEnd{},
	}

	start.End.Start = start

	return start
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindConditionalStart = goldast.NewNodeKind("ConditionalStart")
var KindConditionalEnd = goldast.NewNodeKind("ConditionalEnd")

// Kind implements Node.Kind.
func (n *ConditionalStart) Kind() goldast.NodeKind {
	return KindConditionalStart
}

func (n *ConditionalStart) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"IfScript": n.ifScript, "ID": n.ID}, nil)
}

func (n *ConditionalStart) GetEndSkipNode(goldast.Node) goldast.Node {
	return n.End
}

// Kind implements Node.Kind.
func (n *ConditionalEnd) Kind() goldast.NodeKind {
	return KindConditionalEnd
}

func (n *ConditionalEnd) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"ID": n.Start.ID}, nil)
}

type Conditional interface {
	HasIfScript() bool
	GetIfScript() string
	SetIfScript(string)
	GetResult() bool
	SetResult(bool)
	HasResult() bool
}

type ConditionalImpl struct {
	ifScript       string
	ifScriptResult *bool
}

func (n *ConditionalImpl) HasIfScript() bool {
	return n.ifScript != ""
}

func (n *ConditionalImpl) GetIfScript() string {
	return n.ifScript
}

func (n *ConditionalImpl) SetIfScript(script string) {
	n.ifScript = script
}

func (n *ConditionalImpl) GetResult() bool {
	if n.ifScriptResult != nil {
		return *n.ifScriptResult
	} else {
		return false
	}
}

func (n *ConditionalImpl) HasResult() bool {
	return n.ifScriptResult != nil
}

func (n *ConditionalImpl) SetResult(result bool) {
	n.ifScriptResult = &result
}
