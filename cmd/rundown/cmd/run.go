package cmd

import (
	"errors"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	var callBack func()

	rd, err := rundown.LoadFile(rundownFile)
	if err != nil {
		return err
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	KillReadlineBell()

	switch {

	case flagCompletions != "":

		runCompletions(flagCompletions, cmd)
		return nil

	case flagCodes:

		codes, err := rundown.ParseShortCodeSpecs([]string{"help"})
		if err == nil {
			rd.RunCodes(codes)
		}

		err = RenderShortCodes()
		return handleError(cmd.OutOrStderr(), err, callBack)

	case flagAsk:
		spec, err := AskShortCode()
		if err != nil {
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

		if spec != nil {
			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

	case flagAskRepeat:
		for {
			spec, err := AskShortCode()
			if err != nil {
				return handleError(cmd.OutOrStderr(), err, callBack)
			}

			if spec == nil {
				break
			}

			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			handleError(cmd.OutOrStderr(), err, callBack)
		}

	default:
		specs := argShortcodes

		if len(specs) == 0 && flagDefault != "" {
			specs = []string{flagDefault}
		}

		codes, err := rundown.ParseShortCodeSpecs(specs)
		if err == nil {
			err, callBack = rd.RunCodes(codes)
		}

		if err != nil {
			util.Debugf("ERROR %#v\n", err)
			if errors.Is(err, rundown.InvocationError) {
				codes, err2 := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
				if err2 == nil {
					rd.RunCodesWithoutValidation(codes)
				}

				util.Debugf("Rendering shortcodes for %s\n", rd.Filename())
				err2 = RenderShortCodes()
				util.Debugf("ERROR %#v\n", err2)
			}
		}

		return handleError(cmd.OutOrStderr(), err, callBack)
	}

	return nil
}
