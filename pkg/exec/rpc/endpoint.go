package rpc

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/elseano/rundown/pkg/bus"
	"github.com/elseano/rundown/pkg/util"
)

var EnvironmentVariableName = "RDRPC"

type Endpoint struct {
	Path string
	file *os.File
}

type RpcMessage struct {
	bus.Event
	Data string
}

func Start() (*Endpoint, error) {
	endpoint := &Endpoint{}

	if err := endpoint.initialize(); err != nil {
		return nil, err
	}

	go endpoint.receiveLoop()

	return endpoint, nil
}

func (e *Endpoint) Close() {
	e.file.Close()
	os.Remove(e.Path)
}

func (e *Endpoint) initialize() error {
	file, err := ioutil.TempFile("", "rundown-rpc-*")
	if err != nil {
		return err
	}

	e.Path = file.Name()

	os.Remove(e.Path)

	err = syscall.Mkfifo(e.Path, 0666)
	if err != nil {
		return err
	}

	// RDWR so it doesn't block on opening.
	file, err = os.OpenFile(e.Path, os.O_CREATE|os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return err
	}

	e.file = file

	return nil
}

func (e *Endpoint) receiveLoop() {
	reader := bufio.NewReader(e.file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		trimmedMessage := string(bytes.TrimRight(line, "\r\n"))
		util.Logger.Debug().Msgf("Got message: %s", trimmedMessage)

		bus.Emit(&RpcMessage{Data: trimmedMessage})
	}
}
