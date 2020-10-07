package rundown

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/elseano/rundown/pkg/util"
)

type Context struct {
	Env             map[string]string
	Messages        chan string
	TempDir         string
	ForcedLevelZero bool
	Repeat          bool
	Invocation      string
	ConsoleWidth    int
	Logger          *log.Logger
}

func receiveLoop(filename string, messages chan<- string) {

	os.Remove(filename)
	err := syscall.Mkfifo(filename, 0666)
	if err != nil {
		return
	}

	// RDWR so it doesn't block on opening.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		messages <- string(bytes.TrimRight(line, "\r\n"))
	}
}

func NewContext() *Context {
	messages := make(chan string, 200)

	tmpDir, err := ioutil.TempDir("", "rundown")
	if err != nil {
		panic(err)
	}
	os.MkdirAll(tmpDir, 0644)

	tmpFile, err := ioutil.TempFile(tmpDir, "rpc-*")
	if err != nil {
		panic(err)
	}
	tmpFile.Close()

	go receiveLoop(tmpFile.Name(), messages)

	return &Context{
		Env:             map[string]string{"RUNDOWN": tmpFile.Name()},
		ForcedLevelZero: false,
		Repeat:          false,
		ConsoleWidth:    util.IntMin(util.GetConsoleWidth(), 120),
		Messages:        messages,
		TempDir:         tmpDir,
	}
}

func (c *Context) SetEnvString(envString string) {
	split := strings.SplitN(envString, "=", 2)
	c.Env[split[0]] = split[1]
}

func (c *Context) SetEnv(key, value string) {
	c.Env[key] = value
}

func (c *Context) RemoveEnv(key string) {
	delete(c.Env, key)
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