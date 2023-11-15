package term

import (
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
			message := command[len(changeSpinnerTitleCommand):]

			util.Logger.Debug().Msgf("Set spinner title %s", message)

			spinner.NewStep(message)

		case strings.HasPrefix(command, setEnvironmentCommand):
			_args := command[len(setEnvironmentCommand)-1:]
			args := strings.SplitN(_args, "=", 2)

			util.Logger.Debug().Msgf("Got environment %s = %s", strings.TrimSpace(args[0]), args[1])
			context.Env[strings.TrimSpace(args[0])] = args[1]
		}
	}
}
