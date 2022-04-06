package term

import (
	"encoding/base64"
	"strings"

	"github.com/elseano/rundown/pkg/renderer"
	"github.com/elseano/rundown/pkg/util"
)

func HandleCommands(spinner Spinner, context *renderer.Context) func(s string) {
	const changeSpinnerTitleCommand = "R;SETSPINNER "
	const setEnvironmentCommand = "R;SETENV "

	return func(command string) {
		util.Logger.Debug().Msgf("Got command: %s", command)

		switch {

		case strings.HasPrefix(command, changeSpinnerTitleCommand):
			args := command[len(changeSpinnerTitleCommand):]
			util.Logger.Debug().Msgf("Decoding: %s", args)

			result, err := base64.StdEncoding.DecodeString(args)

			if err == nil {
				util.Logger.Debug().Msgf("Set spinner title %s", result)

				spinner.NewStep(string(result))
			} else {
				util.Logger.Debug().Msgf("Decode err %s", err.Error())
			}

		case strings.HasPrefix(command, setEnvironmentCommand):
			_args := command[len(setEnvironmentCommand)-1:]
			args := strings.SplitN(_args, "=", 2)

			util.Logger.Debug().Msgf("Got environment %s = %s", strings.TrimSpace(args[0]), args[1])
			context.Env[strings.TrimSpace(args[0])] = args[1]

		}
	}
}
