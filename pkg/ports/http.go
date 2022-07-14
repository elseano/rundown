package ports

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/renderer"
	"github.com/elseano/rundown/pkg/transformer"
	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func getDoc(filename string) (goldmark.Markdown, goldast.Node, []byte, *renderer.Context) {
	source, _ := ioutil.ReadFile(filename)

	var buffer bytes.Buffer

	context := renderer.NewContext(filename)
	context.Output = &buffer

	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		context.AddEnv(parts[0], parts[1])
	}

	rundownRenderer := renderer.NewRundownHtmlRenderer(context)
	goldHtml := html.NewRenderer()

	goldmarkRenderer := renderer.NewFlushingRenderer(
		goldrenderer.WithNodeRenderers(
			util.Prioritized(rundownRenderer, 1),
			util.Prioritized(goldHtml, 100),
		),
	)

	// RundownRenderer overrides some Glamour stuff, so we don't need it here.
	// renderer := renderer.NewRundownRenderer(goldmarkRenderer, context)

	gm := goldmark.New(
		goldmark.WithRenderer(goldmarkRenderer),

		goldmark.WithExtensions(extension.GFM),

		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    transformer.NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	return gm, doc, source, context
}

func renderDoc(w http.ResponseWriter, r *http.Request, gm goldmark.Markdown, doc goldast.Node, source []byte, sections []*ast.SectionPointer) {
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html")

	piper, pipew := io.Pipe()

	go func() {
		if err := gm.Renderer().Render(pipew, source, doc); err != nil {
			io.WriteString(w, err.Error())
		} else {
			pipew.Close()
		}
	}()

	lineReader := bufio.NewReader(piper)

	flusher, ok := w.(http.Flusher)

	for {
		fmt.Printf("Reading line...\n")
		line, isPrefix, err := lineReader.ReadLine()
		if err != nil {
			// Write out TOC
			w.Write([]byte("<ul>"))

			for _, s := range sections {
				w.Write([]byte(fmt.Sprintf("<li><a href='/%s'>%s</a></li>", s.SectionName, s.DescriptionShort)))
			}

			w.Write([]byte("</ul>"))

			if ok {
				flusher.Flush()
			}

			return
		}

		fmt.Printf("Sending line: %s\n", line)
		w.Write([]byte(line))
		if !isPrefix {
			w.Write([]byte("\n"))
		}

		if ok {
			flusher.Flush()
		}
	}
}

func strPrt(str string) *string {
	if str == "" {
		return nil
	} else {
		return &str
	}
}

func boolPtr(str string) *bool {
	f := false
	t := true

	if str == "true" {
		return &t
	} else {
		return &f
	}
}

func ServeRundown(filename string, debug bool, port string) error {
	if debug {
		devNull, _ := os.Create("rundown.log")
		rdutil.RedirectLogger(devNull)
	}

	_, err := ioutil.ReadFile(filename)

	fmt.Printf("Serving file %s\n", filename)

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return err
	}

	_, doc, _, _ := getDoc(filename)
	sections := ast.GetSections(doc)

	for _, s := range sections {
		section := s

		http.HandleFunc("/"+section.SectionName, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				gm, doc, source, _ := getDoc(filename)

				doc = ast.PruneDocumentToSection(doc, section.SectionName)
				ast.PruneActions(doc)

				renderDoc(w, r, gm, doc, source, sections)
			} else if r.Method == "POST" {
				gm, doc, source, context := getDoc(filename)

				doc = ast.PruneDocumentToSection(doc, section.SectionName)

				r.ParseForm()

				optionEnv := map[string]optVal{}
				for _, o := range section.Options {
					opt := o
					switch opt.OptionType.(type) {
					case *ast.TypeString:
						optionEnv[opt.OptionAs] = optVal{Str: strPrt(r.FormValue(opt.OptionName)), Option: opt}
					case *ast.TypeBoolean:
						optionEnv[opt.OptionAs] = optVal{Bool: boolPtr(r.FormValue(opt.OptionName)), Option: opt}
					case *ast.TypeEnum:
						optionEnv[opt.OptionAs] = optVal{Str: strPrt(r.FormValue(opt.OptionName)), Option: opt}
					case *ast.TypeFilename:
						optionEnv[opt.OptionAs] = optVal{Str: strPrt(r.FormValue(opt.OptionName)), Option: opt}
					}
				}

				for k, v := range optionEnv {
					if err := v.Option.OptionType.Validate(v.String()); err != nil {
						io.WriteString(w, fmt.Sprintf("%s: %s", v.Option.OptionName, err.Error()))
					}

					context.ImportEnv(map[string]string{
						k: fmt.Sprintf("%v", v.String()),
					})
				}

				renderDoc(w, r, gm, doc, source, sections)
			}
		})
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gm, doc, source, _ := getDoc(filename)
		doc.Dump(source, 0)

		ast.PruneDocumentToRoot(doc)
		ast.PruneActions(doc)

		renderDoc(w, r, gm, doc, source, sections)
	})

	fmt.Printf("Server running on %s", port)

	return http.ListenAndServe(":8080", nil)
}
