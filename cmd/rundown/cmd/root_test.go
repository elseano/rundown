package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elseano/rundown/pkg/rundown"

	"github.com/stretchr/testify/assert"

	"github.com/elseano/rundown/pkg/util"
	"github.com/elseano/rundown/testutil"
)

func TestFullRender(t *testing.T) {
	root, err := filepath.Abs("../../../_testdata/")

	if assert.Nil(t, err) {
		files, err := ioutil.ReadDir(root)
		assert.Nil(t, err)
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".md") {
				t.Run(file.Name(), func(t *testing.T) {
					expected, actual := runSequential(t, path.Join(root, file.Name()))
					testutil.AssertLines(t, expected, actual)
				})
			}

		}
	}
}

// func TestSimpleRundown(t *testing.T) {
// 	fp, _ := filepath.Abs("../../../_testdata/subenv.md")
// 	expected, actual := runSequential(t, fp)

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestSpacingRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/spacing.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestRpcRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/rpc.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestStdoutRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/stdout.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestFailureRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/failure.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestOnFailureRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/on_failure.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestEmojiRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/emoji.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestHiddenRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/hidden.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestFormattingRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/formatting.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestStopOkRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/stop.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// func TestStopFailRundown(t *testing.T) {
// 	expected, actual := runSequential(t, "_testdata/stopfail.md")

// 	testutil.AssertLines(t, expected, actual)
// }

// Blah
func runSequential(t *testing.T, filename string) (string, string) {
	util.Debugf("NEW RUN!\n\n")
	source, expected := loadFile(filename)

	tf, err := ioutil.TempFile(path.Dir(filename), "tmp-"+path.Base(filename)+"-*")
	tempFilename := tf.Name()
	assert.Nil(t, err)
	defer os.Remove(tempFilename)

	tf.WriteString(source)
	tf.Close()
	util.Debugf("Source file is %s\n", tf.Name())
	// util.Debugf("Source contents is %s\n", source)

	var buffer bytes.Buffer

	root := RootCmd()
	root.SetOut(&buffer)
	root.SetErr(&buffer)
	root.PreRun(root, []string{})
	root.ParseFlags([]string{"--cols", "80", "-f", tf.Name()})

	rd, err := rundown.LoadFile(tf.Name())
	util.Debugf("Loading: %s: %#v\n", tf.Name(), err)
	assert.Nil(t, err)
	codes := rd.GetShortCodes()

	fmt.Printf("Load err: %#v\n", err)
	fmt.Printf("Codes: %#v\n", codes.Codes)

	// ast, sourceb := rd.GetAST()
	// util.Debugf("Things %#v\n", sourceb)
	// ast.Dump(sourceb, 0)

	rundownFile = tf.Name()
	util.Debugf("Rundown File = %s\n", rundownFile)

	if codes.Codes["direct"] != nil {
		argShortcodes = []string{"direct"}
		run(root, []string{})
	} else {
		argShortcodes = []string{}
		run(root, []string{})
	}

	fixed := strings.TrimSpace(util.CollapseReturns(util.RemoveColors(buffer.String())))

	t.Logf("Rendering result for %s:", filename)

	for i, line := range strings.Split(fixed, "\n") {
		t.Logf("%3d: %q\n", i, strings.TrimRight(line, " "))
	}

	return expected, fixed
}

func loadFile(absFilename string) (source string, expected string) {
	// absFilename, _ := filepath.Abs("../../../" + filename)
	fp, err := os.Open(absFilename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	scanningSource := true

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "-----") {
			scanningSource = false
		} else if scanningSource {
			source = source + "\n" + line
		} else {
			expected = expected + "\n" + line
		}
	}

	return strings.TrimSpace(source), strings.TrimSpace(expected)
}
