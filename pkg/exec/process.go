package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kr/pty"
	"github.com/thecodeteam/goodbye"
	"golang.org/x/crypto/ssh/terminal"
)

var stdinChannel = make(chan []byte, 500)
var stdinMutex sync.Mutex
var undoRawRegisterd sync.Once
var currentUndoRaw func()

// Reading directly from STDIN when executing multiple processes
// creates issues, as STDIN read is blocking and you can't cancel it.
// Instead, this is the only place we read from STDIN, and we direct the
// byte stream to the io.Reader returned by Claim().
type StdinReader struct {
	init     sync.Once
	stop     chan struct{}
	stopWait sync.WaitGroup
}

func (r *StdinReader) doInit() {
	r.init.Do(func() {
		go func() {
			var buf = make([]byte, 2048)

			for {
				n, err := os.Stdin.Read(buf)

				if err != nil {
					return
				}

				stdinChannel <- buf[0:n]
			}
		}()
	})
}

type stdinReaderCloser struct {
	io.ReadCloser
	reader *StdinReader
}

func (r *stdinReaderCloser) Close() error {
	r.reader.Stop()
	return nil
}

func (r *StdinReader) Claim() io.ReadCloser {
	stdinMutex.Lock()
	r.doInit()

	stdinR, stdinW := io.Pipe()

	r.stopWait.Add(1)

	go func() {
		for {
			select {
			case chars := <-stdinChannel:
				stdinW.Write(chars)
			case <-r.stop:
				stdinW.Close()
				stdinMutex.Unlock()
				r.stopWait.Done()
				return
			}
		}

	}()

	return &stdinReaderCloser{ReadCloser: stdinR, reader: r}
}

func (r *StdinReader) Stop() {
	r.stop <- struct{}{}
	r.stopWait.Wait()
}

func NewStdinReader() *StdinReader {
	reader := &StdinReader{
		stop: make(chan struct{}),
	}

	return reader
}

type Process struct {
	cmd       *exec.Cmd
	pty       *os.File
	stdout    io.Reader
	stdin     *StdinReader
	waitGroup sync.WaitGroup
	mutex     sync.Mutex
	undoRaw   func()
}

func (p *Process) setRawMode() func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, p.pty); err != nil {
				// log.Printf("error resizing pty: %s", err) // Don't care.
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Don't care.
	}

	return func() {
		if oldState != nil {
			terminal.Restore(int(os.Stdin.Fd()), oldState)
		}

		currentUndoRaw = nil
	}

}

func NewProcess(cmd *exec.Cmd) *Process {
	return &Process{cmd: cmd}
}

func (p *Process) Start() (*io.PipeReader, *io.PipeReader, error) {
	p.stdin = NewStdinReader()
	stdinR := p.stdin.Claim()
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	// Reset TTY back if it's not set.
	if currentUndoRaw != nil {
		currentUndoRaw()
	}

	p.undoRaw = p.setRawMode()
	currentUndoRaw = p.undoRaw

	undoRawRegisterd.Do(func() {
		goodbye.Register(func(ctx context.Context, s os.Signal) {
			if currentUndoRaw != nil {
				currentUndoRaw()
			}
		})
	})

	var err error
	if p.pty, err = pty.Start(p.cmd); err != nil {
		return nil, nil, err
	}

	// Copy PTY to output capturing pipe
	go func() {
		defer stdoutR.Close()

		p.waitGroup.Add(1)

		_, _ = io.Copy(stdoutW, p.pty)

		p.waitGroup.Done()
	}()

	go func() {
		defer stderrR.Close()

		p.waitGroup.Add(1)

		_, _ = io.Copy(stderrW, p.pty)

		p.waitGroup.Done()
	}()

	// Copy stdin to PTY
	go func() {
		p.waitGroup.Add(1)

		_, _ = io.Copy(p.pty, stdinR)

		p.waitGroup.Done()

	}()

	// Return output stream reader
	return stdoutR, stderrR, nil
}

func (p *Process) Wait() error {
	defer p.undoRaw()

	err := p.cmd.Wait()
	p.stdin.Stop()
	p.waitGroup.Wait()

	return err
}
