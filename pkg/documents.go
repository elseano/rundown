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

func (doc *LoadedDocument) GetSections() []*ast.SectionPointer {
	result := []*ast.SectionPointer{}

	for child := doc.Document.FirstChild(); child != nil; child = child.NextSibling() {
		if section, ok := child.(*ast.SectionPointer); ok {
			result = append(result, section)
		}
	}

	return result
}

func (d *LoadedDocument) Render(outputStream io.Writer) error {
	return d.Goldmark.Renderer().Render(outputStream, d.Source, d.Document)
}
