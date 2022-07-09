package spinner

import (
	"github.com/elseano/rundown/pkg/util"
)

type wrappableSpinner interface {
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
	StampShadow()
}

type SubenvSpinner struct {
	env     map[string]string
	spinner wrappableSpinner
}

func NewSubenvSpinner(env map[string]string, spinner wrappableSpinner) *SubenvSpinner {
	return &SubenvSpinner{
		env:     env,
		spinner: spinner,
	}
}

func (s *SubenvSpinner) Active() bool {
	return s.spinner.Active()
}

func (s *SubenvSpinner) Start() {
	s.spinner.Start()
}

func (s *SubenvSpinner) Stop() {
	s.spinner.Stop()
}

func (s *SubenvSpinner) Success(message string) {
	message = s.SubEnv(message)

	s.spinner.Success(message)
}

func (s *SubenvSpinner) Error(message string) {
	message = s.SubEnv(message)
	s.spinner.Error(message)
}

func (s *SubenvSpinner) Skip(message string) {
	message = s.SubEnv(message)
	s.spinner.Skip(message)
}

func (s *SubenvSpinner) SetMessage(message string) {
	message = s.SubEnv(message)
	s.spinner.SetMessage(message)
}

func (s *SubenvSpinner) NewStep(message string) {
	message = s.SubEnv(message)
	s.spinner.NewStep(message)
}

func (s *SubenvSpinner) HideAndExecute(f func()) {
	s.spinner.HideAndExecute(f)
}

func (s *SubenvSpinner) CurrentHeading() string {
	return s.spinner.CurrentHeading()
}

func (s *SubenvSpinner) StampShadow() {
	s.spinner.StampShadow()
}

func (s *SubenvSpinner) SubEnv(message string) string {
	return util.SubEnv(s.env, message)
}
