package segments

import (
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/elseano/rundown/testutil"

	"github.com/yuin/goldmark"

	"github.com/stretchr/testify/assert"
)

func TestPrepareSegments(t *testing.T) {
	contents := []byte("Normal markdown text\n\n<!--~\nHidden markdown text, only for rundown\n-->\nMore text")

	markdown := markdown.PrepareMarkdown()

	logger := testutil.NewTestLogger(t)
	segments := BuildSegments(string(contents), markdown, logger)

	logger.Printf("%v", segments)

	var result string

	assert.Equal(t, "DisplaySegment", segments[0].Kind())

	result = renderSeg(segments[0], markdown, contents, logger)
	assert.Equal(t, "Normal markdown text\n\nHidden markdown text, only for rundown\n\nMore text", result)

}

func TestPrepareSegmentsSpacing(t *testing.T) {
	contents := []byte("Normal markdown text\n\n``` bash\necho 'Hi'\n```\n\nMore Text.")

	markdown := markdown.PrepareMarkdown()

	logger := log.New(os.Stdout, "DEBUG: ", log.Ltime)
	segments := BuildSegments(string(contents), markdown, logger)

	log.Printf("%v", segments)

	assert.Equal(t, "DisplaySegment", segments[0].Kind())
	assert.Equal(t, "CodeSegment", segments[1].Kind())
	assert.Equal(t, "Separator", segments[2].Kind())
	assert.Equal(t, "DisplaySegment", segments[3].Kind())
}

func renderSeg(seg Segment, md goldmark.Markdown, contents []byte, logger *log.Logger) string {
	context := &Context{
		Env: map[string]string{
			"RUNDOWN": "",
		},
		Messages:     make(chan string),
		TempDir:      "",
		ConsoleWidth: 80,
	}

	rx := regexp.MustCompile("\033\\[[0-9\\;]+m")

	trimmed := strings.TrimSpace(util.CaptureStdout(func() { seg.Execute(context, md.Renderer(), nil, logger, os.Stdout) }))
	cleaned := rx.ReplaceAllString(trimmed, "")

	return util.CollapseReturns(util.RemoveColors(cleaned))
}
