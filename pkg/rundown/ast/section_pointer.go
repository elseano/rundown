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
	DescriptionLong  string
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
