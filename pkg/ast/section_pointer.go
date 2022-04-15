package ast

import (
	"fmt"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark/ast"
	goldast "github.com/yuin/goldmark/ast"
)

type SectionPointer struct {
	goldast.BaseBlock
	ConditionalImpl

	SectionName      string
	StartNode        goldast.Node
	Options          []*SectionOption
	DescriptionShort string
	DescriptionLong  *DescriptionBlock
	Silent           bool

	Dependencies []*SectionPointer
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionPointer(name string) *SectionPointer {
	p := &SectionPointer{
		SectionName: name,
		Options:     []*SectionOption{},
	}

	return p
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionPointer = goldast.NewNodeKind("SectionPointer")

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

func (n *SectionPointer) FirstContentNode() goldast.Node {
	for node := n.NextSibling(); node != nil; node = node.NextSibling() {
		if node.Kind() != goldast.KindHeading {
			return node
		}
	}

	return nil
}

func (n *SectionPointer) GetOption(name string) *SectionOption {
	for _, o := range n.Options {
		if o.OptionName == name {
			return o
		}
	}

	return nil
}

func (n *SectionPointer) ParseOptions(options map[string]string) (map[string]string, error) {
	return n.ParseOptionsWithResolution(options, map[string]string{})
}

func (n *SectionPointer) ParseOptionsWithResolution(options map[string]string, env map[string]string) (map[string]string, error) {
	result := map[string]string{}

	util.Logger.Debug().Msgf("Env is %+v", env)

	for k, v := range options {
		option := n.GetOption(k)

		if option != nil {
			optionValue := option.OptionType.Normalise(v)

			if rt, ok := option.OptionType.(OptionTypeRuntime); ok {
				wd, err := os.Getwd()
				if err != nil {
					return nil, err
				}

				ov, err := rt.NormaliseToPath(v, wd)

				if err != nil {
					return nil, err
				}

				optionValue = ov
			}

			if err := option.OptionType.Validate(optionValue); err != nil {
				return nil, fmt.Errorf("%s: %w", option.OptionName, err)
			}

			result[option.OptionAs] = fmt.Sprintf("%v", optionValue)

			for k, v := range env {
				result[option.OptionAs] = strings.Replace(result[option.OptionAs], "$"+k, v, -1)
			}
		}
	}

	return result, nil
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

	goldast.Walk(doc, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if entering {
			if section, ok := n.(*SectionPointer); ok {
				result = append(result, section)
			}
		}

		return ast.WalkContinue, nil
	})

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

// // Removes everything from the current node through to either SectionEnd or no more nodes.
// func PruneSectionFromNode(node goldast.Node) {

// 	// Walk through all following siblings of the current node and delete them.

// 	sib := node.NextSibling()
// 	for sib != nil {
// 		if _, ok := sib.(*SectionEnd); ok {
// 			return
// 		}

// 		nextSib := sib.NextSibling()
// 		sib.Parent().RemoveChild(sib.Parent(), sib)
// 		sib = nextSib
// 	}

// 	// The do the same for the node's parent, unless it's a document.
// 	parent := node.Parent()

// 	if _, ok := parent.(*goldast.Document); ok {
// 		return
// 	}

// 	PruneSectionFromNode(parent)
// }

// Reduces the document to just the requested section.
func PruneDocumentToSection(doc goldast.Node, sectionName string) {
	var sectionPointer *SectionPointer = FindSectionInDocument(doc, sectionName)

	for child := doc.FirstChild(); child != nil; {
		nextChild := child.NextSibling()

		if child != sectionPointer {
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
