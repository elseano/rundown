package segments

import (
	"strings"
	"regexp"
	"errors"
)

type Flag string

const (
	NoSpinFlag            Flag = "nospin"
	InteractiveFlag            = "interactive"
	SkipOnSuccessFlag          = "skip_on_success"
	SkipOnFailureFlag          = "skip_on_failure"
	StdoutFlag                 = "stdout"
	StderrFlag                 = "stderr"
	RevealFlag                 = "reveal"
	NoRunFlag                  = "norun"
	NamedFlag                  = "named"
	NamedAllFlag 							 = "named_all"
	CaptureEnvFlag					   = "env"
	AbortFlag					         = "abort"
	IgnoreFailureFlag					 = "ignore_failure"
	EnvAwareFlag					     = "env_aware"
	SetupFlag					         = "setup"
)

var validFlags = map[Flag]bool{ 
	NoSpinFlag: true, 
	InteractiveFlag: true, 
	SkipOnSuccessFlag: true, 
	SkipOnFailureFlag: true, 
	StdoutFlag: true, 
	StderrFlag: true, 
	RevealFlag: true, 
	NoRunFlag: true,
	NamedFlag: true,
	NamedAllFlag: true, 
	CaptureEnvFlag: true, 
	AbortFlag: true, 
	IgnoreFailureFlag: true, 
	EnvAwareFlag: true, 
	SetupFlag: true,
}

type Parameter string

const (
	SaveParameter Parameter = "save"
	WithParameter           = "with"
	LabelParameter          = "label"
)

var validParameters = map[Parameter]bool{ 
	SaveParameter: true, 
	WithParameter: true,
	LabelParameter: true,
}

type Modifiers struct {
	Values map[Parameter]string
	Flags map[Flag]bool
}

func NewModifiers() *Modifiers {
	return &Modifiers{
		Values: map[Parameter]string{},
		Flags: map[Flag]bool{},
	}
}

func (m *Modifiers) Ingest(from *Modifiers) {
	for k,v := range from.Flags {
		m.Flags[k] = v
	}

	for k,v := range from.Values {
		m.Values[k] = v
	}
}

var modifierDetect = regexp.MustCompile("([a-z_]+\\:\".*?\")\\s*|([a-z_]+\\:'.*?')\\s*|([a-z_]+\\:[^\\s]+)\\s*|([a-z0-9_]+)\\s*")

const (
	quotedKV = 1
	quotedSingleKV = 2
	unquotedKV = 3
	flagV = 4
)

func ParseModifiers(line string) *Modifiers {

	result := NewModifiers()

	for _, match := range modifierDetect.FindAllStringSubmatch(line, -1) {
		if subject := match[quotedKV]; subject != "" {
			if kv := strings.SplitN(subject, ":", 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = strings.Trim(kv[1], "\"")
			}
		} else if subject := match[quotedSingleKV]; subject != "" {
			if kv := strings.SplitN(subject, ":", 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = strings.Trim(kv[1], "'")
			}
		} else if subject := match[unquotedKV]; subject != "" {
			if kv := strings.SplitN(subject, ":", 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = kv[1]
			}
		} else if subject := match[flagV]; subject != "" {
			result.Flags[Flag(subject)] = true
		}
	}

	return result
}

func MergeModifiers(left, right *Modifiers) *Modifiers {
	result := NewModifiers()
	result.Ingest(left)
	result.Ingest(right)
	return result
}

func ValidateModifiers(subject *Modifiers) []error {
	result := []error{}

	for key, _ := range subject.Flags {
		if _, ok := validFlags[key]; !ok {
			result = append(result, errors.New(string(key)))
		}
	}

	for key, _ := range subject.Values {
		if _, ok := validParameters[key]; !ok {
			result = append(result, errors.New(string(key)))
		}
	}

	return result
}
