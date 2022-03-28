package spinner

import (
	"fmt"
	"io"
	"time"

	"github.com/logrusorgru/aurora"
)

type StdoutSpinner struct {
	s              *ActualSpinner
	colorMode      aurora.Aurora
	message        string
	out            io.Writer
	lastStampTitle string
}

const (
	TICK  = "✔"
	CROSS = "✖"
	SKIP  = "≡"
)

func NewStdoutSpinner(colorMode aurora.Aurora, colorsEnabled bool, out io.Writer) *StdoutSpinner {
	opts := []Option{WithWriter(out)}
	if colorsEnabled {
		opts = append(opts, WithColor("fgHiCyan"))
	}

	s := NewActualSpinner(CharSets[21], 100*time.Millisecond, opts...)
	return &StdoutSpinner{s: s, out: out, colorMode: colorMode}
}

func (s *StdoutSpinner) Active() bool {
	return s.s.Active()
}

func (s *StdoutSpinner) Spin() {
	s.Start()
}

func (s *StdoutSpinner) Start() {
	s.s.Start()
}

func (s *StdoutSpinner) StampShadow() {
	if s.lastStampTitle != s.message {
		fmt.Fprintf(s.out, "%s\r\n", s.colorMode.Faint(fmt.Sprintf("↓ %s", s.message)))
		s.lastStampTitle = s.message
	}
}

func (s *StdoutSpinner) Success(message string) {
	s.s.FinalMSG = s.colorMode.Green(TICK).String() + " " + s.message + " (" + s.colorMode.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *StdoutSpinner) Error(message string) {
	s.s.FinalMSG = s.colorMode.Red(CROSS).String() + " " + s.message + " (" + s.colorMode.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *StdoutSpinner) Skip(message string) {
	s.s.FinalMSG = s.colorMode.Faint(SKIP).String() + " " + s.message + " (" + s.colorMode.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *StdoutSpinner) Stop() {
	s.s.Stop()
}

func (s *StdoutSpinner) CurrentHeading() string {
	return s.message
}

func (s *StdoutSpinner) NewStep(message string) {
	var wasActive = s.Active()
	s.Success("OK")

	sp := NewActualSpinner(CharSets[21], 100*time.Millisecond, WithWriter(s.out), WithColor("fgHiCyan"))
	sp.Suffix = " " + message

	s.message = message
	s.s = sp

	if wasActive {
		s.Start()
	}
}

func (s *StdoutSpinner) SetMessage(message string) {
	s.message = message
	s.s.Suffix = " " + message //+ "\033[1A"

	if s.s.Active() {
		s.s.Repaint()
	}
}

func (s *StdoutSpinner) HideAndExecute(f func()) {
	s.s.HideAndExecute(f)
}
