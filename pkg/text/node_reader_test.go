package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	// "bytes"
	// "io/ioutil"
)

func TestReadSingleSegmentFull(t *testing.T) {
	reader := text.NewReader([]byte("Text"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 4))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 4)

	nodeReader := NewNodeReader(node, reader)
	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, buffer, []byte("Text"))
	assert.Equal(t, n, 4)
}

func TestReadSingleSegmentUnder(t *testing.T) {
	reader := text.NewReader([]byte("TextAndMore"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 11))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 4)

	nodeReader := NewNodeReader(node, reader)
	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, buffer, []byte("Text"))
	assert.Equal(t, n, 4)
}

func TestReadSingleSegmentOver(t *testing.T) {
	reader := text.NewReader([]byte("Text"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 4))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 14)

	nodeReader := NewNodeReader(node, reader)
	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, buffer[0:4], []byte("Text"))
	assert.Equal(t, n, 4)

}

func TestReadMultipleSegmentFull(t *testing.T) {
	reader := text.NewReader([]byte("TextMore"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 4))
	segments.Append(text.NewSegment(4, 8))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 4)

	nodeReader := NewNodeReader(node, reader)

	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, []byte("Text"), buffer)
	assert.Equal(t, n, 4)

	n, _ = nodeReader.Read(buffer)

	assert.EqualValues(t, "More", string(buffer))
	assert.Equal(t, n, 4)
}

func TestReadMultipleSegmentUnder(t *testing.T) {
	reader := text.NewReader([]byte("TextAndMore"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 7))
	segments.Append(text.NewSegment(7, 11))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 4)

	nodeReader := NewNodeReader(node, reader)

	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, "Text", string(buffer))
	assert.Equal(t, n, 4)

	n, _ = nodeReader.Read(buffer)

	assert.Equal(t, "And", string(buffer[0:3]))
	assert.Equal(t, n, 3)

	n, _ = nodeReader.Read(buffer)

	assert.EqualValues(t, "More", string(buffer))
	assert.Equal(t, n, 4)
}

func TestReadMultipleSegmentOver(t *testing.T) {
	reader := text.NewReader([]byte("TextAndMore"))

	segments := text.NewSegments()
	segments.Append(text.NewSegment(0, 7))
	segments.Append(text.NewSegment(7, 11))

	node := goldast.NewDocument()
	node.SetLines(segments)

	var buffer = make([]byte, 40)

	nodeReader := NewNodeReader(node, reader)

	n, _ := nodeReader.Read(buffer)

	assert.Equal(t, "TextAnd", string(buffer[0:7]))
	assert.Equal(t, n, 7)

	n, _ = nodeReader.Read(buffer)

	assert.EqualValues(t, "More", string(buffer[0:4]))
	assert.Equal(t, n, 4)
}
