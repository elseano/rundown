package rundown

import (
	"errors"

	"github.com/elseano/rundown/pkg/markdown"
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
	OnFailureFlag                   = "on-failure"
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
	OnFailureFlag:     true,
}

const (
	SaveParameter        markdown.Parameter = "save"
	WithParameter                           = "with"
	LabelParameter                          = "label"
	DescriptionParameter                    = "desc"
	OnFailureParameter                      = "on-failure"
	StopFailParameter                       = "stop-fail"
)

var validParameters = map[markdown.Parameter]bool{
	SaveParameter:        true,
	WithParameter:        true,
	LabelParameter:       true,
	DescriptionParameter: true,
	OnFailureParameter:   true,
	StopFailParameter:    true,
}

func ValidateModifiers(subject *markdown.Modifiers) []error {
	result := []error{}

	for key := range subject.Flags {
		if _, ok := validFlags[key]; !ok {
			result = append(result, errors.New(string(key)))
		}
	}

	for key := range subject.Values {
		if _, ok := validParameters[key]; !ok {
			result = append(result, errors.New(string(key)))
		}
	}

	return result
}
