package exec

import (
	"fmt"
	"testing"

	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorParseBash(t *testing.T) {

	sm, err := scripts.NewScript("bash", "bash", []byte("one\ntwo\nthree\nfour\nfive\nsix"))
	require.NoError(t, err)

	output := fmt.Sprintf("%s: line 5: DOCKER_BUILD_TAG: unbound variable", sm.AbsolutePath)

	parsed := ParseError(sm, output)

	if assert.NotNil(t, parsed.ErrorSource) {
		if assert.Equal(t, 5, parsed.ErrorSource.Line) {
			assert.Equal(t, "DOCKER_BUILD_TAG: unbound variable", parsed.Error)
		}
	}
}

func TestErrorParseBashAdjustsForPrefix(t *testing.T) {

	sm, err := scripts.NewScript("bash", "bash", []byte("one\ntwo\nthree\nfour\nfive\nsix"))
	sm.Prefix = []byte("Some\nPre\nLines\n")
	require.NoError(t, err)

	output := fmt.Sprintf("%s: line 5: DOCKER_BUILD_TAG: unbound variable", sm.AbsolutePath)

	parsed := ParseError(sm, output)

	if assert.NotNil(t, parsed.ErrorSource) {
		if assert.Equal(t, 2, parsed.ErrorSource.Line) {
			assert.Equal(t, "DOCKER_BUILD_TAG: unbound variable", parsed.Error)
		}
	}
}
