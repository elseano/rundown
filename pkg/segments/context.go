package segments

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type Context struct {
	Env             map[string]string
	Messages        chan string
	TempDir         string
	ForcedLevelZero bool
	Repeat          bool
	Invocation      string
	ConsoleWidth    int
}

func intMin(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func getConsoleWidth() int {
	width, _, err := terminal.GetSize(0)
	if err != nil {
		width = 80
	}

	return width
}

func NewContext() *Context {
	return &Context{
		Env:             map[string]string{},
		ForcedLevelZero: false,
		Repeat:          false,
		ConsoleWidth:    intMin(getConsoleWidth(), 120),
	}
}

func (c *Context) SetEnvString(envString string) {
	split := strings.SplitN(envString, "=", 2)
	c.Env[split[0]] = split[1]
}

func (c *Context) SetEnv(key, value string) {
	c.Env[key] = value
}

func (c *Context) EnvStringList() []string {
	var result = []string{}

	for k, v := range c.Env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result
}

var envMatch = regexp.MustCompile("(\\$[A-Z0-9_]+)")

func SubEnv(content string, context *Context) (string, error) {

	for k, v := range context.Env {
		content = strings.ReplaceAll(content, "$"+k, v)
	}

	if match := envMatch.FindString(content); match != "" {
		return content, errors.New(match + " is not set")
	}

	return content, nil

}
