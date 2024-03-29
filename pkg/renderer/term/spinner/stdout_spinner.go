package spinner

import (
	"fmt"
	"io"
	"time"

	"github.com/elseano/rundown/pkg/util"
	"github.com/logrusorgru/aurora"
)

type StdoutSpinner struct {
	s              *ActualSpinner
	colorMode      aurora.Aurora
	message        string
	substep        string
	out            io.Writer
	lastStampTitle string
	startedAt      time.Time
}

const (
	TICK  = "✔"
	CROSS = "✖"
	SKIP  = "⋯"
	DASH  = "-"
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
	if s.startedAt.IsZero() {
		s.startedAt = time.Now()
	}

	util.Logger.Debug().Msg("Starting spinner...")

	s.s.Start()
}

func (s *StdoutSpinner) StampShadow() {
	if s.substep != "" && s.lastStampTitle != s.substep {
		fmt.Fprintf(s.out, "%s\r\n", s.colorMode.Faint(fmt.Sprintf("  ↓ %s", s.substep)))
		s.lastStampTitle = s.substep
	} else if s.substep == "" && s.lastStampTitle != s.message {
		fmt.Fprintf(s.out, "%s\r\n", s.colorMode.Faint(fmt.Sprintf("↓ %s", s.message)))
		s.lastStampTitle = s.message
	}
}

func (s *StdoutSpinner) closeSpinner(indicator string, message string, faint bool) {
	if message != "" {
		message = " " + message
	}

	colour := s.colorMode.Black

	if faint {
		colour = func(arg any) aurora.Value { return s.colorMode.Faint(s.colorMode.StrikeThrough(arg)) }
	}

	ts := buildTimeString(s.startedAt)

	if s.substep != "" {

		s.s.FinalMSG = "  " + indicator + " " + colour(s.substep+message).String() + " " + s.colorMode.Faint(ts).String() + "\r\n"
		s.Stop()

		sp := NewActualSpinner(CharSets[21], 100*time.Millisecond, WithWriter(s.out), WithColor("fgHiCyan"))
		sp.Suffix = " " + s.message
		sp.Start()
		sp.FinalMSG = indicator + " " + s.message + message + "\r\n"
		sp.Stop()
	} else {
		s.s.FinalMSG = indicator + " " + colour(s.message+message).String() + " " + s.colorMode.Faint(ts).String() + "\r\n"
		s.Stop()
	}

}

func (s *StdoutSpinner) Success(message string) {
	s.closeSpinner(s.colorMode.Green(TICK).String(), message, false)
}

func (s *StdoutSpinner) Error(message string) {
	s.closeSpinner(s.colorMode.Red(CROSS).String(), message, false)
}

func (s *StdoutSpinner) Skip() {
	s.closeSpinner(s.colorMode.Faint(SKIP).String(), "", true)
}

func (s *StdoutSpinner) Stop() {
	s.s.Stop()
}

func (s *StdoutSpinner) CurrentHeading() string {
	return s.message
}

func (s *StdoutSpinner) NewStep(message string) {
	var wasActive = s.Active()

	util.Logger.Debug().Msgf("NewStep spinner was active: %v", wasActive)

	ts := time.Since(s.startedAt).String()

	if s.substep == "" {
		util.Logger.Debug().Msgf("Dangling heading...")
		s.s.FinalMSG = s.colorMode.Faint(DASH).String() + " " + s.message + "\r\n"
		s.Stop()
	} else {
		util.Logger.Debug().Msgf("Closing step...")
		s.s.FinalMSG = "  " + s.colorMode.Green(TICK).String() + " " + s.substep + s.colorMode.Faint(" ("+ts+")").String() + "\r\n"
		s.Stop()
	}

	util.Logger.Debug().Msgf("Creating new spinner...")

	sp := NewActualSpinner(CharSets[21], 100*time.Millisecond, WithWriter(s.out), WithColor("fgHiCyan"))
	sp.Suffix = " " + message
	sp.Prefix = "  "

	s.substep = message
	s.s = sp
	s.startedAt = time.Time{}

	util.Logger.Debug().Msgf("Done")

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
