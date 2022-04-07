package spinner

import (
	"fmt"
	"io"
	"time"

	"github.com/elseano/rundown/pkg/util"
	"github.com/logrusorgru/aurora"
)

type section struct {
	codeName string
	title    string
}

type GitlabSpinner struct {
	out            io.Writer
	section        []section
	sectionPrefix  string
	sectionCounter int
	colors         aurora.Aurora
}

func NewGitlabSpinner(out io.Writer, colors aurora.Aurora) *GitlabSpinner {
	return &GitlabSpinner{
		out:            out,
		sectionPrefix:  util.RandomString(),
		sectionCounter: 0,
		section:        []section{},
		colors:         colors,
	}
}

func (s *GitlabSpinner) Active() bool {
	return false
}

func (s *GitlabSpinner) Start() {}

func (s *GitlabSpinner) Stop() {}

func (s *GitlabSpinner) StampShadow() {
	// No need, as we've opened a section already during the SetMessage call.
}

func (s *GitlabSpinner) closeSpinner(indicator string) {
	for len(s.section) > 0 {
		if indicator == CROSS {
			currentSection := s.section[len(s.section)-1]
			s.out.Write([]byte(fmt.Sprintf("%s %s", indicator, currentSection.title)))
		}

		s.closeSection()
	}
}

func (s *GitlabSpinner) Success(message string) {
	s.closeSpinner(TICK)
}

func (s *GitlabSpinner) Error(message string) {
	s.closeSpinner(CROSS)
}

func (s *GitlabSpinner) Skip(message string) {
	s.closeSpinner(SKIP)
}

func (s *GitlabSpinner) SetMessage(message string) {
	s.closeSection()
	s.openSection(message)
}

func (s *GitlabSpinner) NewStep(message string) {
	if len(s.section) > 1 {
		s.closeSection()
	}

	s.openSection(message)
}

func (s *GitlabSpinner) HideAndExecute(f func()) {

}

func (s *GitlabSpinner) CurrentHeading() string {
	return ""
}

func (s *GitlabSpinner) closeSection() {
	if len(s.section) > 0 {
		currentSection := s.section[len(s.section)-1]
		nowStr := time.Now().Unix()
		s.out.Write([]byte(fmt.Sprintf("\n\033[0Ksection_end:%d:%s\r\033[0K\r\n", nowStr, currentSection.codeName)))
		s.section = s.section[0 : len(s.section)-1]
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *GitlabSpinner) openSection(name string) {
	s.sectionCounter++
	currentSection := fmt.Sprintf("sec_%s_%d", s.sectionPrefix, s.sectionCounter)
	s.section = append(s.section, section{codeName: currentSection, title: name})
	nowStr := time.Now().Unix()
	s.out.Write([]byte(fmt.Sprintf("\033[0Ksection_start:%d:%s\r\033[0K%s\r\n", nowStr, currentSection, s.colors.BrightCyan(name).String())))
}
