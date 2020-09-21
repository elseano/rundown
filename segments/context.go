package segments

import "strings"

type Context struct {
	Env      map[string]string
	Messages chan string
	TempDir  string
	ForcedIndentZero bool
}

func (c *Context) SetEnvString(envString string) {
	split := strings.SplitN(envString, "=", 2)
	c.Env[split[0]] = split[1]
}

func (c *Context) SetEnv(key, value string) {
	c.Env[key] = value
}
