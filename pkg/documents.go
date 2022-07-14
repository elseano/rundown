package rundown

import (
	"io"

	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/renderer"
	"github.com/yuin/goldmark"
	goldast "github.com/yuin/goldmark/ast"
)

type LoadedDocuments struct {
	MasterDocument    *LoadedDocument
	ImportedDocuments []*LoadedDocument
	Context           *renderer.Context
}

type Section struct {
	Pointer  *ast.SectionPointer
	Document *LoadedDocument
}

func (doc *LoadedDocuments) GetSections() []*Section {
	result := []*Section{}

	for _, section := range doc.MasterDocument.GetSections() {
		result = append(result, &Section{
			Pointer:  section,
			Document: doc.MasterDocument,
		})
	}

	for _, d := range doc.ImportedDocuments {
		for _, section := range d.GetSections() {
			result = append(result, &Section{
				Pointer:  section,
				Document: d,
			})
		}
	}

	return result
}

type LoadedDocument struct {
	Filename string
	Document goldast.Node
	Source   []byte
	Goldmark goldmark.Markdown
	Context  *renderer.Context
}

// Walks through the document and returns all the found SectionPointers
func (doc *LoadedDocument) GetSections() []*ast.SectionPointer {
	result := []*ast.SectionPointer{}

	goldast.Walk(doc.Document, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if section, ok := n.(*ast.SectionPointer); entering && ok {
			result = append(result, section)
		}

		return goldast.WalkContinue, nil
	})

	return result
}

// Walks through the document and returns all the found InvokeBlocks
func (doc *LoadedDocument) GetInvokes() []*ast.InvokeBlock {
	result := []*ast.InvokeBlock{}

	goldast.Walk(doc.Document, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if section, ok := n.(*ast.InvokeBlock); entering && ok {
			result = append(result, section)
		}

		return goldast.WalkContinue, nil
	})

	return result
}

func (d *LoadedDocument) Render(outputStream io.Writer) error {
	return d.Goldmark.Renderer().Render(outputStream, d.Source, d.Document)
}
