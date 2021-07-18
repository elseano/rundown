package cmd

import (
	"bufio"
	"bytes"
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

// Test a specific file via:
//
// go test ./... -run TestRunFile/subenv.md

func TestRunFile(t *testing.T) {
	root, err := filepath.Abs("../../../_testdata/")

	if assert.Nil(t, err) {
		files, err := ioutil.ReadDir(root)
		assert.Nil(t, err)
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".md") && !strings.HasPrefix(file.Name(), "custom-") {
				t.Run(file.Name(), func(t *testing.T) {
					expected, actual := runSequential(t, path.Join(root, file.Name()))
					testutil.AssertLines(t, expected, actual)
				})
			}

		}
	}
}

func TestRunFileCustomOptions(t *testing.T) {
	file, err := filepath.Abs("../../../_testdata/custom-opt-parse.md")
	if err != nil {
		t.Error(err)
	}

	expected, actual := runSequential(t, file, "options", "+name=Blah")
	testutil.AssertLines(t, expected, actual)
}

func TestRunFileCustomInvalidOptions(t *testing.T) {
	file, err := filepath.Abs("../../../_testdata/custom-opt-parse.md")
	if err != nil {
		t.Error(err)
	}

	_, actual := runSequential(t, file, "options", "+incorrect=Blah")
	testutil.AssertLines(t, "Error: Option 'options+incorrect' is not supported", actual)
}

func TestRunFileCustomOptionsEnum(t *testing.T) {
	file, err := filepath.Abs("../../../_testdata/custom-opt-parse.md")
	if err != nil {
		t.Error(err)
	}

	_, actual := runSequential(t, file, "options:enum", "+thing=yep")
	testutil.AssertLines(t, "Enum\n\n  Result: yep", actual)
}

func TestRunFileCustomInvalidOptionsEnum(t *testing.T) {
	file, err := filepath.Abs("../../../_testdata/custom-opt-parse.md")
	if err != nil {
		t.Error(err)
	}

	_, actual := runSequential(t, file, "options:enum", "+thing=nope")
	testutil.AssertLines(t, "Error: Option 'options:enum+thing' must be one of: yep, nah", actual)
}

// Handy test of a single file for debugging

func TestDebug(t *testing.T) {
	file, err := filepath.Abs("../../../_testdata/on_failure.md")
	if err != nil {
		t.Error(err)
	}

	expected, actual := runSequential(t, file)
	testutil.AssertLines(t, expected, actual)
}

func runSequential(t *testing.T, filename string, extraArgs ...string) (string, string) {
	source, expected := loadFile(filename)

	tf, err := ioutil.TempFile(path.Dir(filename), "tmp-"+path.Base(filename)+"-*")
	tempFilename := tf.Name()
	assert.Nil(t, err)
	defer os.Remove(tempFilename)

	tf.WriteString(source)
	tf.Close()
	t.Logf("Source file is %s\n", tf.Name())
	t.Logf("Source contents is %s\n", source)

	var buffer bytes.Buffer

	root := NewRootCmd()
	root.SetOut(&buffer)
	root.SetErr(&buffer)

	rd, err := rundown.LoadFile(tf.Name())
	util.Debugf("Loading: %s: %#v\n", tf.Name(), err)
	assert.Nil(t, err)
	codes := rd.GetShortCodes()

	t.Logf("Load err: %#v\n", err)
	t.Logf("Codes: %#v\n", codes.Codes)

	ast, src := rd.GetAST()
	t.Log(util.DumpNode(ast, src))

	if _, ok := codes.Codes["direct"]; ok {
		t.Logf("Rundown file has a direct section. Running that instead.\n")
		root.SetArgs(append([]string{"--cols", "80", "-f", tf.Name(), "direct"}, extraArgs...))
	} else {
		t.Logf("Rundown file contains no direct section, running whole file.\n")
		root.SetArgs(append([]string{"--cols", "80", "-f", tf.Name()}, extraArgs...))
	}

	root.Execute()

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
