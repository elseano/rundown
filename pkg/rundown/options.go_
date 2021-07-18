package rundown

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/util"
)

func validateOptionValue(specs map[string]*ShortCodeOptionSpec, opt *ShortCodeOption) string {
	_, isSet := specs[opt.Code]

	// Inject the default value
	if opt.Default != "" && !isSet {
		specs[opt.Code] = &ShortCodeOptionSpec{Code: opt.Code, Value: opt.Default}
	}

	var value string

	if spec, ok := specs[opt.Code]; ok {
		value = spec.Value
	} else {
		return "is required"
	}

	if opt.Required && value == "" {
		if opt.Prompt {
			return "" // Will be prompted later.
		}

		return "is required"
	}

	switch opt.Type {
	case "string":
		return "" // Strings are always just fine.
	case "file-exists":
		if !util.FileExists(value) {
			return "file not found"
		}

		return ""
	case "int":
		fallthrough
	case "integer":
		fallthrough
	case "number":
		_, err := strconv.ParseInt(value, 10, 0)

		if err != nil {
			return err.Error()
		}

		return ""

	case "file-not-exists":
		if util.FileExists(value) {
			return "file already exists"
		}
		return ""

	case "bool":
		fallthrough
	case "boolean":
		v, err := strconv.ParseBool(value)

		if err != nil {
			return err.Error()
		}

		// Noramlize it

		if v {
			specs[opt.Code].Value = "TRUE"
		} else {
			specs[opt.Code].Value = "FALSE"
		}

		return ""
	default:

		if strings.HasPrefix(opt.Type, "enum|") {
			possibles := strings.Split(opt.Type, "|")[1:]

			for _, poss := range possibles {
				if value == poss {
					return ""
				}
			}

			return fmt.Sprintf("must be one of: %s", strings.Join(possibles, ", "))
		}

	}

	return ""

}

func ValidateOptions(docSpec *DocumentSpec, shortCodes *DocumentShortCodes) error {

	// fmt.Printf("Validating %s\n", docSpec)

	if len(shortCodes.Codes) == 0 && len(docSpec.ShortCodes) > 0 {
		return &InvalidShortCodeError{ShortCode: docSpec.ShortCodes[0].Code}
	}

	if len(shortCodes.Options) == 0 && len(docSpec.Options) > 0 {
		for key := range docSpec.Options {
			return &InvalidOptionsError{OptionName: key, Message: "is not supported"}
		}
	}

	for optName := range docSpec.Options {
		opt := shortCodes.Options[optName]
		if opt == nil {
			return &InvalidOptionsError{OptionName: optName, Message: "is not supported"}
		}
	}

	for _, opt := range shortCodes.Options {
		if err := validateOptionValue(docSpec.Options, opt); err != "" {
			return &InvalidOptionsError{OptionName: opt.Code, Message: err}
		}
	}

	for _, code := range docSpec.ShortCodes {
		section := shortCodes.Codes[code.Code]
		if section == nil {
			return &InvalidShortCodeError{ShortCode: code.Code}
		}

		for _, opt := range code.Options {
			if section.Options[opt.Code] == nil {
				return &InvalidOptionsError{OptionName: opt.Code, ShortCode: code.Code, Message: "is not supported"}
			}
		}

		for _, opt := range section.Options {
			if err := validateOptionValue(code.Options, opt); err != "" {
				return &InvalidOptionsError{OptionName: opt.Code, ShortCode: code.Code, Message: err}
			}
		}
	}

	return nil
}
