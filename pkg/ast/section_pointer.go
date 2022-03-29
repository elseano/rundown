package ast

import (
	"strings"

	goldast "github.com/yuin/goldmark/ast"
)

type SectionPointer struct {
	goldast.BaseBlock
	SectionName      string
	StartNode        goldast.Node
	Options          []*SectionOption
	DescriptionShort string
	DescriptionLong  *DescriptionBlock
}

type SectionEnd struct {
	goldast.BaseBlock
	SectionPointer *SectionPointer
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionPointer(name string) *SectionPointer {
	return &SectionPointer{
		SectionName: name,
		Options:     []*SectionOption{},
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionPointer = goldast.NewNodeKind("SectionPointer")
var KindSectionEnd = goldast.NewNodeKind("SectionEnd")

// Kind implements Node.Kind.
func (n *SectionPointer) Kind() goldast.NodeKind {
	return KindSectionPointer
}

func (n *SectionPointer) Dump(source []byte, level int) {
	optionNames := []string{}
	for _, opt := range n.Options {
		optionNames = append(optionNames, opt.OptionName)
	}

	goldast.DumpHelper(n, source, level, map[string]string{"SectionName": n.SectionName, "Options": strings.Join(optionNames, ", ")}, nil)
}

func (n *SectionPointer) AddOption(option *SectionOption) {
	n.Options = append(n.Options, option)
}

// Kind implements Node.Kind.
func (n *SectionEnd) Kind() goldast.NodeKind {
	return KindSectionEnd
}

func (n *SectionEnd) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"SectionName": n.SectionPointer.SectionName}, nil)
}

func FindSectionInDocument(parent goldast.Node, name string) *SectionPointer {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if section, ok := child.(*SectionPointer); ok {
			if section.SectionName == name {
				return section
			}
		}
	}

	return nil
}

func GetSections(doc goldast.Node) []*SectionPointer {
	result := []*SectionPointer{}

	for child := doc.FirstChild(); child != nil; {
		nextChild := child.NextSibling()

		if section, ok := child.(*SectionPointer); ok {
			result = append(result, section)
		}

		child = nextChild
	}

	return result
}

func GetSectionForNode(node goldast.Node) *SectionPointer {
	// First, get the containing block element.
	for node.Parent().Kind() != goldast.KindDocument {
		node = node.Parent()
	}

	// The walk backwards until we find the SectionPointer
	for node != nil && node.Kind() != KindSectionPointer {
		node = node.PreviousSibling()
	}

	if p, ok := node.(*SectionPointer); ok {
		return p
	}

	return nil
}

func PruneDocumentToRoot(doc goldast.Node) {
	rootEnded := false

	for child := doc.FirstChild(); child != nil; {
		nextChild := child.NextSibling()

		if _, ok := child.(*SectionPointer); ok {
			rootEnded = true
		}

		if rootEnded {
			doc.RemoveChild(doc, child)
		}

		child = nextChild
	}
}

// Removes everything from the current node through to either SectionEnd or no more nodes.
func PruneSectionFromNode(node goldast.Node) {

	// Walk through all following siblings of the current node and delete them.

	sib := node.NextSibling()
	for sib != nil {
		if _, ok := sib.(*SectionEnd); ok {
			return
		}

		nextSib := sib.NextSibling()
		sib.Parent().RemoveChild(sib.Parent(), sib)
		sib = nextSib
	}

	// The do the same for the node's parent, unless it's a document.
	parent := node.Parent()

	if _, ok := parent.(*goldast.Document); ok {
		return
	}

	PruneSectionFromNode(parent)
}

// Reduces the document to just the requested section.
func PruneDocumentToSection(doc goldast.Node, sectionName string) {
	var sectionPointer *SectionPointer = nil

	for child := doc.FirstChild(); child != nil; {
		nextChild := child.NextSibling()

		if section, ok := child.(*SectionPointer); ok {
			if section.SectionName == sectionName {
				sectionPointer = section
			}
		} else if sectionEnd, ok := child.(*SectionEnd); ok {
			if sectionEnd.SectionPointer == sectionPointer {
				sectionPointer = nil
			}
		}

		if sectionPointer == nil {
			doc.RemoveChild(doc, child)
		}

		child = nextChild
	}

}

func PruneActions(doc goldast.Node) {
	actionsFound := false

	for child := doc.FirstChild(); child != nil; {
		nextChild := child.NextSibling()

		if _, ok := child.(*ExecutionBlock); ok {
			actionsFound = true
		}

		if actionsFound {
			doc.RemoveChild(doc, child)
		}

		child = nextChild
	}
}
