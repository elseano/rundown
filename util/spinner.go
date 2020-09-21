package util

import (
	"strings"
	"time"
	"io"

	// spx "github.com/tj/go-spin"
	// spx "github.com/briandowns/spinner"
	"github.com/elseano/rundown/spinner"
	"github.com/logrusorgru/aurora"
)

const (
	TICK  = "✔"
	CROSS = "✖"
	SKIP  = "≡"
)

type Spinner interface {
	Active() bool
	Start()
	Stop()
	Success(message string)
	Error(message string)
	Skip(message string)
	SetMessage(message string)
	HideAndExecute(f func())
}

type RundownSpinner struct {
	s       *spinner.ActualSpinner
	indent  int
	message string
}

func NewSpinner(indent int, message string, out io.Writer) Spinner {
	s := spinner.NewActualSpinner(spinner.CharSets[21], 100*time.Millisecond, spinner.WithWriter(out))
	s.Suffix = " " + message
	s.Prefix = strings.Repeat("  ", indent)

	return &RundownSpinner{indent: indent, message: message, s: s}
}

func (s *RundownSpinner) Active() bool {
	return s.s.Active()
}

func (s *RundownSpinner) Spin() {
	s.Start()
}

func (s *RundownSpinner) Start() {
	s.s.Start()
}

func (s *RundownSpinner) Success(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + aurora.Green(TICK).String() + " " + s.message + " (" + aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Error(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + aurora.Red(CROSS).String() + " " + s.message + " (" + aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Skip(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + aurora.Faint(SKIP).String() + " " + s.message + " (" + aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Stop() {
	s.s.Stop()
}

func (s *RundownSpinner) SetMessage(message string) {
	s.message = message
	s.s.Suffix = " " + message //+ "\033[1A"
	s.s.Repaint()
}

func (s *RundownSpinner) HideAndExecute(f func()) {
	s.s.HideAndExecute(f)
}

type DummySpinner struct {
	active bool
}


func NewDummySpinner() Spinner {
	return &DummySpinner{ active: false }
}

func (s *DummySpinner) Active() bool {
	return s.active
}

func (s *DummySpinner) Spin() {
	s.Start()
}

func (s *DummySpinner) Start() {
}

func (s *DummySpinner) Success(message string) {
	s.Stop()
}

func (s *DummySpinner) Error(message string) {
	s.Stop()
}

func (s *DummySpinner) Skip(message string) {
	s.Stop()
}

func (s *DummySpinner) Stop() {
}

func (s *DummySpinner) SetMessage(message string) {
}

func (s *DummySpinner) HideAndExecute(f func()) {
	f()
}
