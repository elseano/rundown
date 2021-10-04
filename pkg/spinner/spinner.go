package spinner

import (
	"io"
	"strings"
	"time"

	// spx "github.com/tj/go-spin"
	// spx "github.com/briandowns/spinner"

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
	NewStep(message string)
	HideAndExecute(f func())
	CurrentHeading() string
}

type RundownSpinner struct {
	s       *ActualSpinner
	indent  int
	message string
	out     io.Writer
}

func NewSpinner(indent int, message string, out io.Writer) Spinner {
	s := NewActualSpinner(CharSets[21], 100*time.Millisecond, WithWriter(out), WithColor("fgHiCyan"))
	s.Suffix = " " + message
	s.Prefix = strings.Repeat("  ", indent)

	return &RundownSpinner{indent: indent, message: message, s: s, out: out}
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

func (s *RundownSpinner) CurrentHeading() string {
	return s.message
}

func (s *RundownSpinner) NewStep(message string) {
	var wasActive = s.Active()
	s.Success("OK")

	sp := NewActualSpinner(CharSets[21], 100*time.Millisecond, WithWriter(s.out), WithColor("fgHiCyan"))
	sp.Suffix = " " + message
	sp.Prefix = strings.Repeat("  ", s.indent)

	s.message = message
	s.s = sp

	if wasActive {
		s.Start()
	}
}

func (s *RundownSpinner) SetMessage(message string) {
	s.message = message
	s.s.Suffix = " " + message //+ "\033[1A"

	if s.s.Active() {
		s.s.Repaint()
	}
}

func (s *RundownSpinner) HideAndExecute(f func()) {
	s.s.HideAndExecute(f)
}

type DummySpinner struct {
	active bool
}

func NewDummySpinner() Spinner {
	return &DummySpinner{active: false}
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

func (s *DummySpinner) NewStep(message string) {
}

func (s *DummySpinner) SetMessage(message string) {
}

func (s *DummySpinner) HideAndExecute(f func()) {
	f()
}

func (s *DummySpinner) CurrentHeading() string {
	return ""
}
