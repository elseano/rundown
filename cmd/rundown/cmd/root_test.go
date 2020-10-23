package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			ok := true

			if !info.IsDir() && filepath.Ext(path) == ".md" {
				filename := filepath.Base(path)
				ok = t.Run(filename, func(t *testing.T) {
					expected, actual := runSequential(t, path)
					testutil.AssertLines(t, expected, actual)
				})
			}

			if !ok {
				return errors.New("Stop")
			}

			return nil
		})
	}
}

func TestSimpleRundown(t *testing.T) {
	fp, _ := filepath.Abs("../../../_testdata/subenv.md")
	expected, actual := runSequential(t, fp)

	testutil.AssertLines(t, expected, actual)
}

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
	source, expected := loadFile(filename)

	tf, err := ioutil.TempFile("", "")
	assert.Nil(t, err)

	tf.WriteString(source)
	tf.Close()
	defer os.Remove(tf.Name())

	var buffer bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{tf.Name()})
	root.SetOut(&buffer)
	root.SetErr(&buffer)
	root.PreRun(root, []string{tf.Name()})
	root.ParseFlags([]string{"--cols", "80"})

	rd, err := rundown.LoadFile(tf.Name())
	codes := rd.GetShortCodes()

	fmt.Printf("Codes: %#v", codes.Codes)

	if codes.Codes["direct"] != nil {
		argShortcodes = []string{"direct"}
		run(root, []string{tf.Name()})
	} else {
		argShortcodes = []string{}
		run(root, []string{tf.Name()})
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
