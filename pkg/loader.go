package rundown

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/elseano/rundown/pkg/ast"
	rundown_parser "github.com/elseano/rundown/pkg/parser"
	"github.com/elseano/rundown/pkg/renderer"
	"github.com/elseano/rundown/pkg/transformer"
	emoji "github.com/yuin/goldmark-emoji"

	termrend "github.com/elseano/rundown/pkg/renderer/term"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type LoadErrors []error

func (le *LoadErrors) Error() string {
	s := strings.Builder{}
	for _, e := range *le {
		s.WriteString(fmt.Sprintf("%s\n", e.Error()))
	}

	return s.String()
}

func LoadString(data string, filename string) (*LoadedDocuments, error) {
	context := renderer.NewContext(filename)

	parentDocument, err := loadBytes([]byte(data), filename, context)

	if err != nil {
		return nil, err
	}

	return cascadeLoad(parentDocument)
}

func Load(filename string) (*LoadedDocuments, error) {
	context := renderer.NewContext(filename)
	parentDocument, err := loadFile(filename, context)

	if err != nil {
		return nil, err
	}

	return cascadeLoad(parentDocument)
}

func cascadeLoad(parentDocument *LoadedDocument) (*LoadedDocuments, error) {
	collection := &LoadedDocuments{
		MasterDocument:    parentDocument,
		ImportedDocuments: []*LoadedDocument{},
		Context:           parentDocument.Context,
	}

	currentPath := path.Dir(parentDocument.Filename)
	importDirectives := ast.ProcessImportBlocks(parentDocument.Document)

	for _, directive := range importDirectives {
		filename := directive.GetFilename()

		importedDoc, err := loadFile(path.Join(currentPath, filename), parentDocument.Context)
		if err != nil {
			return nil, err
		}

		// If we have an import prefix, prepend it to the section names
		if directive.ImportPrefix != "" {
			for _, section := range importedDoc.GetSections() {
				section.SectionName = fmt.Sprintf("%s:%s", directive.ImportPrefix, section.SectionName)
			}

			for _, invoke := range importedDoc.GetInvokes() {
				invoke.Invoke = fmt.Sprintf("%s:%s", directive.ImportPrefix, invoke.Invoke)
			}
		}

		collection.ImportedDocuments = append(collection.ImportedDocuments, importedDoc)
	}

	return collection, nil

}

func loadFile(filename string, context *renderer.Context) (*LoadedDocument, error) {
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, &LoadErrors{err}
	}

	return loadBytes(source, filename, context)
}

func loadBytes(source []byte, filename string, context *renderer.Context) (*LoadedDocument, error) {
	consoleNodeRenderer := termrend.NewRenderer(context)
	renderer := goldrenderer.WithNodeRenderers(
		util.Prioritized(consoleNodeRenderer, 0),
	)

	rdtransform := transformer.NewRundownASTTransformer()

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    rdtransform,
				Priority: 0,
			}),
		),
		goldmark.WithRendererOptions(renderer),
		goldmark.WithExtensions(emoji.New(), rundown_parser.InvisibleBlocks),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	if len(rdtransform.Errors) > 0 {
		errs := LoadErrors(rdtransform.Errors)
		return nil, &errs
	}

	return &LoadedDocument{
		Filename: filename,
		Document: doc,
		Source:   source,
		Goldmark: gm,
		Context:  context,
	}, nil
}
