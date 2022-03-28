package scripts

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/elseano/rundown/pkg/util"
)

type ScriptManager struct {
	scripts         map[string]*Script
	lastScriptAdded *Script
}

func NewScriptManager() *ScriptManager {
	return &ScriptManager{
		scripts: map[string]*Script{},
	}
}

func (m *ScriptManager) SetBaseScript(invocation string, contents []byte) (*Script, error) {
	return m.AddScript("BASE", invocation, contents)
}

func (m *ScriptManager) AllScripts() []*Script {
	result := []*Script{}
	for _, s := range m.scripts {
		result = append(result, s)
	}

	return result
}

func (m *ScriptManager) AddScript(name, invocation string, contents []byte) (*Script, error) {
	script := &Script{
		AbsolutePath:     "",
		Contents:         contents,
		Invocation:       invocation,
		EnvReferenceName: strings.ToUpper("SCRIPT_" + name),
		Name:             name,
		ShellScript:      isShellLike(invocation),
	}

	if _, exists := m.scripts[name]; exists {
		return nil, errors.New("Script with alias " + name + " already exists.")
	}

	m.scripts[name] = script
	m.lastScriptAdded = script

	return script, nil
}

func (m *ScriptManager) GenerateReferences() map[string]string {
	var results = map[string]string{}

	for _, script := range m.scripts {
		if script.AbsolutePath != "" {
			results[script.EnvReferenceName] = script.AbsolutePath
		}
	}

	return results
}

// func (m *ScriptManager) Get(name string) *Script {
// 	return m.scripts[name]
// }

func (m *ScriptManager) GetBase() *Script {
	return m.scripts["BASE"]
}

func (m *ScriptManager) GetPrevious() *Script {
	return m.lastScriptAdded
}

func (m *ScriptManager) Write() (*Script, error) {
	for _, script := range m.scripts {
		if err := script.Write(); err != nil {
			return nil, err
		}
	}

	return m.lastScriptAdded, nil
}

func (m *ScriptManager) RemoveAll() {
	// for _, script := range m.scripts {
	// 	if script.AbsolutePath != "" {
	// 		os.Remove(script.AbsolutePath)
	// 	}
	// }
}

type InterpreterNotFound struct{}

func (e *InterpreterNotFound) Error() string {
	return "Interpreter not found."
}

func buildShebang(via string) ([]byte, error) {
	if via == "" {
		via = "bash"
	}

	parts := strings.SplitN(via, " ", 2)
	execActual := parts[0]

	if filepath.IsAbs(execActual) {
		if !util.FileExists(execActual) {
			return []byte{}, &InterpreterNotFound{}
		}

		return []byte("/usr/bin/env " + via), nil
	}

	path, err := exec.LookPath(via)
	if err != nil {
		return []byte{}, err
	}

	return []byte("/usr/bin/env " + path), nil
}

func prepareScript(shebang []byte, script []byte) (string, error) {
	tmpFile, err := ioutil.TempFile("", "rundown-exec-*")

	if err != nil {
		return "", err
	}

	defer tmpFile.Close()
	filename := tmpFile.Name()

	tmpFile.Write([]byte("#!"))
	tmpFile.Write(shebang)
	tmpFile.Write([]byte("\n\n"))
	tmpFile.Write(script)
	os.Chmod(filename, 0700)

	return filename, nil
}
