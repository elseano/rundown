package renderer

import (
	"io"
	"strings"
)

type Context struct {
	Env         map[string]string
	Output      io.Writer
	RundownFile string
}

func NewContext(rundownFile string) *Context {
	return &Context{
		Env:         map[string]string{},
		RundownFile: rundownFile,
	}
}

func (c *Context) ImportEnv(env map[string]string) {
	for k, v := range env {
		c.Env[k] = v
	}
}

func (c *Context) ImportRawEnv(env []string) {
	for _, v := range env {
		parts := strings.SplitN(v, "=", 2)

		if len(parts) == 2 {
			c.Env[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			c.Env[parts[0]] = ""
		}
	}
}

func (c *Context) AddEnv(key string, value string) {
	c.Env[key] = value
}
