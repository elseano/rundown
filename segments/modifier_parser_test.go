package segments

import (
	"testing"

	"github.com/stretchr/testify/assert"

)

func TestParseFlags(t *testing.T) {
	result := ParseModifiers("simple flags")

	assert.True(t, result.Flags[Flag("simple")])
	assert.True(t, result.Flags[Flag("flags")])
}


func TestParseParams(t *testing.T) {
	result := ParseModifiers("some:Value flags")

	assert.Equal(t, "Value", result.Values[Parameter("some")])
	assert.True(t, result.Flags[Flag("flags")])
}

func TestParseQuotedParams(t *testing.T) {
	result := ParseModifiers("some:\"Value with spaces\" flags")

	assert.Equal(t, "Value with spaces", result.Values[Parameter("some")])
	assert.True(t, result.Flags[Flag("flags")])
}

func TestParseSingleQuotedParams(t *testing.T) {
	result := ParseModifiers("some:'Value with spaces' flags")

	assert.Equal(t, "Value with spaces", result.Values[Parameter("some")])
	assert.True(t, result.Flags[Flag("flags")])
}

func TestParseMixedParams(t *testing.T) {
	result := ParseModifiers("truth some:\"Value with spaces\" flags save:somewhere_again.jpg hide:!not")

	assert.Equal(t, "Value with spaces", result.Values[Parameter("some")])
	assert.Equal(t, "somewhere_again.jpg", result.Values[Parameter("save")])
	assert.Equal(t, "!not", result.Values[Parameter("hide")])
	assert.True(t, result.Flags[Flag("truth")])
	assert.True(t, result.Flags[Flag("flags")])
}
