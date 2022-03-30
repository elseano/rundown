package renderer

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
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

// Opens a temporary file, and adds it's filename to the context environment.
func (c *Context) CreateTempFile(name string) (*os.File, error) {
	nameParts := strings.SplitN(name, ".", 2)
	if len(nameParts) == 1 {
		nameParts = append(nameParts, "")
	}

	file, err := ioutil.TempFile("", fmt.Sprintf("rd-%s-*.%s", nameParts[0], nameParts[1]))
	if err != nil {
		return nil, err
	}

	envName := strings.ToUpper(nameParts[0])
	envName = regexp.MustCompile(`[^A-Z0-9_]`).ReplaceAllString(envName, "_")
	envName = fmt.Sprintf("%s_FILE", envName)

	c.Env[envName] = file.Name()

	return file, nil
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
