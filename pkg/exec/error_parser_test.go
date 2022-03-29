package exec

import (
	"testing"

	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/stretchr/testify/assert"
)

func TestErrorParseBash(t *testing.T) {
	output := "/var/folders/9y/pdzh2ryj5dj0w96wq08bt3l40000gn/T/rd-base-507761123: line 5: DOCKER_BUILD_TAG: unbound variable"

	sm := scripts.NewScriptManager()
	sm.SetBaseScript("bash", []byte("one\ntwo\nthree\nfour\nfive\nsix"))
	sm.AllScripts()[0].AbsolutePath = "/var/folders/9y/pdzh2ryj5dj0w96wq08bt3l40000gn/T/rd-base-507761123"

	err := ParseError(sm, output)

	if assert.NotNil(t, err.ErrorSource) {
		if assert.Equal(t, 4, err.ErrorSource.Line) {
			assert.Equal(t, "DOCKER_BUILD_TAG: unbound variable", err.Error)
		}
	}
}

func TestErrorParseBashAdjustsForPrefix(t *testing.T) {
	output := "/var/folders/9y/pdzh2ryj5dj0w96wq08bt3l40000gn/T/rd-base-507761123: line 5: DOCKER_BUILD_TAG: unbound variable"

	sm := scripts.NewScriptManager()
	script, _ := sm.SetBaseScript("bash", []byte("one\ntwo\nthree\nfour\nfive\nsix"))
	script.AbsolutePath = "/var/folders/9y/pdzh2ryj5dj0w96wq08bt3l40000gn/T/rd-base-507761123"
	script.Prefix = []byte("Some\nPre\nLines")
	err := ParseError(sm, output)

	if assert.NotNil(t, err.ErrorSource) {
		if assert.Equal(t, 0, err.ErrorSource.Line) {
			assert.Equal(t, "DOCKER_BUILD_TAG: unbound variable", err.Error)
		}
	}
}
