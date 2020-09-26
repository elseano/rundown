package segments

import (
	"errors"

	"github.com/elseano/rundown/markdown"
)

const (
	NoSpinFlag        markdown.Flag = "nospin"
	InteractiveFlag                 = "interactive"
	SkipOnSuccessFlag               = "skip-on-success"
	SkipOnFailureFlag               = "skip-on-failure"
	StdoutFlag                      = "stdout"
	StderrFlag                      = "stderr"
	RevealFlag                      = "reveal"
	NoRunFlag                       = "norun"
	NamedFlag                       = "named"
	NamedAllFlag                    = "named-all"
	CaptureEnvFlag                  = "env"
	StopOkFlag                      = "stop-ok"
	StopFailFlag                    = "stop-fail"
	IgnoreFailureFlag               = "ignore-failure"
	EnvAwareFlag                    = "sub-env"
	SetupFlag                       = "setup"
	BorgFlag                        = "borg"
	DescriptionFlag                 = "desc"
)

var validFlags = map[markdown.Flag]bool{
	NoSpinFlag:        true,
	InteractiveFlag:   true,
	SkipOnSuccessFlag: true,
	SkipOnFailureFlag: true,
	StdoutFlag:        true,
	StderrFlag:        true,
	RevealFlag:        true,
	NoRunFlag:         true,
	NamedFlag:         true,
	NamedAllFlag:      true,
	CaptureEnvFlag:    true,
	StopOkFlag:        true,
	StopFailFlag:      true,
	IgnoreFailureFlag: true,
	EnvAwareFlag:      true,
	SetupFlag:         true,
	BorgFlag:          true,
	DescriptionFlag:   true,
}

const (
	SaveParameter        markdown.Parameter = "save"
	WithParameter                           = "with"
	LabelParameter                          = "label"
	DescriptionParameter                    = "desc"
)

var validParameters = map[markdown.Parameter]bool{
	SaveParameter:        true,
	WithParameter:        true,
	LabelParameter:       true,
	DescriptionParameter: true,
}

func ValidateModifiers(subject *markdown.Modifiers) []error {
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
