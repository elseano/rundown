package spinner

import (
	"fmt"
	"io"
	"time"
)

type GitlabSpinner struct {
	out            io.Writer
	section        string
	sectionCounter int
	currentHeading string
}

func NewGitlabSpinner(out io.Writer) *GitlabSpinner {
	return &GitlabSpinner{out: out}
}

func (s *GitlabSpinner) Active() bool {
	return false
}

func (s *GitlabSpinner) Start() {}

func (s *GitlabSpinner) Stop() {}

func (s *GitlabSpinner) StampShadow() {
}

func (s *GitlabSpinner) Success(message string) {
	s.closeSection()
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", TICK, s.CurrentHeading(), message)))
}

func (s *GitlabSpinner) Error(message string) {
	s.closeSection()
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", CROSS, s.CurrentHeading(), message)))
}

func (s *GitlabSpinner) Skip(message string) {
	s.closeSection()
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", SKIP, s.CurrentHeading(), message)))
}

func (s *GitlabSpinner) SetMessage(message string) {
	s.closeSection()
	s.openSection(message)
}

func (s *GitlabSpinner) NewStep(message string) {
	s.SetMessage(message)
}

func (s *GitlabSpinner) HideAndExecute(f func()) {

}

func (s *GitlabSpinner) CurrentHeading() string {
	return s.currentHeading
}

func (s *GitlabSpinner) closeSection() {
	if s.section != "" {
		nowStr := time.Now().Unix()
		s.out.Write([]byte(fmt.Sprintf("\033[0Ksection_end:%d:%s\r\033[0K\r\n", nowStr, s.section)))
		s.section = ""
		s.currentHeading = ""
	}
}

func (s *GitlabSpinner) openSection(name string) {
	s.sectionCounter++
	s.section = fmt.Sprintf("sec%d", s.sectionCounter)
	nowStr := time.Now().Unix()
	s.out.Write([]byte(fmt.Sprintf("\033[0Ksection_start:%d:%s\r\033[0K%s\r\n", nowStr, s.section, name)))
	s.currentHeading = name
}
