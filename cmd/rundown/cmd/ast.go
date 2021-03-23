package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/yuin/goldmark/text"
)

func runAst(filename string) int {
	md := rundown.PrepareMarkdown()

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Errorf("Error: %s", err.Error())
		return 1
	}

	reader := text.NewReader(b)

	doc := md.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	doc.Dump(b, 0)
	return 0
}
