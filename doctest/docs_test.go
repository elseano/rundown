package doctest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/renderer/term"
	"github.com/elseano/rundown/pkg/renderer/term/spinner"
	"github.com/elseano/rundown/pkg/text"
	"github.com/elseano/rundown/pkg/util"
	"github.com/elseano/rundown/testutil"
	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gold_ast "github.com/yuin/goldmark/ast"
)

func TestDocumentation(t *testing.T) {
	devNull, _ := os.Create(os.DevNull)
	util.RedirectLogger(devNull)

	r, err := rundown.Load("../docs/index.md")

	assert.NoError(t, err)

	sections := r.GetSections()

	util.RedirectLogger(testutil.NewTestWriter(t))
	t.Logf("Sections: %#v", sections)

	for _, s := range sections {
		t.Run(s.Pointer.SectionName, func(t *testing.T) {
			util.RedirectLogger(devNull)

			// Reload document as we mutate it.
			r, err := rundown.Load("../docs/index.md")
			assert.NoError(t, err)

			var section *rundown.Section

			// Find the section.
			for _, ss := range r.GetSections() {
				if ss.Pointer.SectionName == s.Pointer.SectionName {
					section = ss
					break
				}
			}

			ast.PruneDocumentToSection(section.Document.Document, section.Pointer.SectionName)

			// Get markdown section
			source := ast.FindNode(section.Document.Document, func(n gold_ast.Node) bool {
				if fcb, ok := n.(*gold_ast.FencedCodeBlock); ok && fcb.Info != nil {
					return string(fcb.Info.Text(section.Document.Source)) == "markdown"
				}

				return false
			})

			// Get expected section
			expected := ast.FindNode(section.Document.Document, func(n gold_ast.Node) bool {
				if fcb, ok := n.(*gold_ast.FencedCodeBlock); ok && fcb.Info != nil {
					info := string(fcb.Info.Text(section.Document.Source))

					return info == "expected" || info == "expected-err"
				}

				return false
			})

			// See if there's a specific invocation mentioned.
			invocation := ast.FindNodeBackwardsDeeply(expected, func(n gold_ast.Node) bool {
				if code, ok := n.(*gold_ast.CodeSpan); ok {
					return strings.Contains(string(code.Text(section.Document.Source)), "rundown ")
				}

				return false
			})

			if assert.NotNil(t, source, "Source is nil") && assert.NotNil(t, expected, "Expected is nil") {
				util.RedirectLogger(testutil.NewTestWriter(t))

				fcbSource := source.(*gold_ast.FencedCodeBlock)
				fcbExpected := expected.(*gold_ast.FencedCodeBlock)

				r := text.NewNodeReaderFromSource(fcbSource, section.Document.Source)
				sourceText, _ := ioutil.ReadAll(r)

				r = text.NewNodeReaderFromSource(fcbExpected, section.Document.Source)
				expectedText, _ := ioutil.ReadAll(r)

				if assert.NotEmpty(t, sourceText) && assert.NotEmpty(t, expectedText) {
					rd, err := rundown.LoadString(string(sourceText), section.Document.Filename)

					output := bytes.Buffer{}

					if assert.NoError(t, err) {
						term.Aurora = aurora.NewAurora(false)
						term.ColorsEnabled = false
						term.NewSpinnerFunc = func(w io.Writer) term.Spinner {
							return NewDocTestSpinner(w)
						}

						require.NoError(t, ast.FillInvokeBlocks(rd.MasterDocument.Document, 10))

						if invocation != nil {
							invocationStr := string(invocation.Text(section.Document.Source))
							sectionName := strings.Replace(invocationStr, "rundown ", "", 1)

							t.Logf("Executing section %s", sectionName)

							ast.PruneDocumentToSection(rd.MasterDocument.Document, sectionName)
						}

						out := util.CaptureStdout(func() {
							rd.MasterDocument.Document.Dump(sourceText, 0)
						})

						t.Log("Document executed:\n")
						t.Log(out)

						err := rd.MasterDocument.Goldmark.Renderer().Render(&output, sourceText, rd.MasterDocument.Document)

						info := string(fcbExpected.Info.Text(section.Document.Source))

						t.Log("Result:\n")
						t.Log(output.String())

						if (info == "expected" && assert.NoError(t, err)) || (info == "expected-err" && assert.Error(t, err)) {
							assert.Equal(t, cleanString(string(expectedText)), cleanString(output.String()))
						}
					}
				}

			}

		})
	}
}

func cleanString(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.TrimSpace(s)
	return s
}

type DocTestSpinner struct {
	w          io.Writer
	m          string
	mLastStamp string
	substep    string
}

func NewDocTestSpinner(w io.Writer) *DocTestSpinner {
	return &DocTestSpinner{w: w}
}

func (s *DocTestSpinner) Active() bool {
	return false
}

func (s *DocTestSpinner) Start() {

}

func (s *DocTestSpinner) Stop() {

}

func (s *DocTestSpinner) Success(message string) {
	if s.substep != "" {
		s.w.Write([]byte(fmt.Sprintf("  %s %s\n", spinner.TICK, s.substep)))
	}

	s.w.Write([]byte(fmt.Sprintf("%s %s\n", spinner.TICK, s.m)))
}

func (s *DocTestSpinner) Error(message string) {
	s.w.Write([]byte(fmt.Sprintf("%s %s\n", spinner.CROSS, s.m)))
}

func (s *DocTestSpinner) Skip(message string) {
	s.w.Write([]byte(fmt.Sprintf("%s %s\n", spinner.SKIP, s.m)))
}

func (s *DocTestSpinner) SetMessage(message string) {
	s.m = message
}

func (s *DocTestSpinner) NewStep(message string) {
	if s.substep != "" {
		s.w.Write([]byte(fmt.Sprintf("  %s %s\n", spinner.TICK, s.substep)))
	} else {
		s.w.Write([]byte(fmt.Sprintf("%s %s\n", spinner.DASH, s.m)))
	}

	s.substep = message
}

func (s *DocTestSpinner) HideAndExecute(f func()) {

}

func (s *DocTestSpinner) CurrentHeading() string {
	return s.m
}

func (s *DocTestSpinner) StampShadow() {
	if s.mLastStamp != s.m {
		s.w.Write([]byte(fmt.Sprintf("↓ %s\n", s.m)))
		s.mLastStamp = s.m
	}
}
