package rundown

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/renderer"
	emoji "github.com/yuin/goldmark-emoji"

	termrend "github.com/elseano/rundown/pkg/rundown/renderer/term"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func Load(filename string) (*LoadedDocuments, error) {
	context := renderer.NewContext(filename)
	parentDocument, err := loadFile(filename, context)

	collection := &LoadedDocuments{
		MasterDocument:    parentDocument,
		ImportedDocuments: []*LoadedDocument{},
		Context:           context,
	}

	if err != nil {
		return nil, err
	}

	currentPath := path.Dir(filename)
	importDirectives := ast.ProcessImportBlocks(parentDocument.Document)

	for _, directive := range importDirectives {
		filename := directive.GetFilename()

		importedDoc, err := loadFile(path.Join(currentPath, filename), context)
		if err != nil {
			return nil, err
		}

		// If we have an import prefix, prepend it to the section names
		if directive.ImportPrefix != "" {
			for _, section := range importedDoc.GetSections() {
				section.SectionName = fmt.Sprintf("%s:%s", directive.ImportPrefix, section.SectionName)
			}
		}

		collection.ImportedDocuments = append(collection.ImportedDocuments, importedDoc)
	}

	return collection, nil
}

func loadFile(filename string, context *renderer.Context) (*LoadedDocument, error) {
	fmt.Printf("Loading %s\n", filename)
	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	consoleNodeRenderer := termrend.NewRenderer(context)
	renderer := goldrenderer.WithNodeRenderers(
		util.Prioritized(consoleNodeRenderer, 0),
	)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    transformer.NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
		goldmark.WithRendererOptions(renderer),
		goldmark.WithExtensions(emoji.New()),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	return &LoadedDocument{
		Filename: filename,
		Document: doc,
		Source:   source,
		Goldmark: gm,
		Context:  context,
	}, nil
}
