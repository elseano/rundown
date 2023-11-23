package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubEnv(t *testing.T) {
	env := map[string]string{
		"CI":            "true",
		"CI_BRANCH":     "blah",
		"CIRCLE_BRANCH": "more",
	}

	require.Equal(t, "Test blah", SubEnv(env, "Test $CI_BRANCH"))
	require.Equal(t, "Test more", SubEnv(env, "Test $CIRCLE_BRANCH"))
	require.Equal(t, "Test blah", SubEnv(env, "Test ${CI_BRANCH}"))
	require.Equal(t, "Test more", SubEnv(env, "Test ${CIRCLE_BRANCH}"))
	require.Equal(t, "Test something", SubEnv(env, "Test ${NOPE:-something}"))
	require.Equal(t, "Test something", SubEnv(env, "Test ${NOPE-something}"))
	require.Equal(t, "Test ", SubEnv(env, "Test ${NOPE+something}"))
	require.Equal(t, "Test something", SubEnv(env, "Test ${CI+something}"))
	require.Equal(t, "Test true", SubEnv(env, "Test ${CI%%something}"))

	require.Equal(t, "Test , when blah... is `:-f` or more", SubEnv(env, "Test ${UNSET+something}, when $CI_BRANCH... is `$NOPE:-f` or ${CIRCLE_BRANCH:-blah}"))
}
