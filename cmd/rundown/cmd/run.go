package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

func buildDocSpec(command string, runner *rundown.Runner, cmd *cobra.Command, args []string) (*rundown.DocumentSpec, error) {
	docShortCodes := runner.GetShortCodes()

	commandCode := docShortCodes.Codes[command]

	if commandCode == nil {
		return nil, fmt.Errorf("invalid section %s", command)
	}

	commandSpec := &rundown.ShortCodeSpec{
		Code:    commandCode.Code,
		Options: map[string]*rundown.ShortCodeOptionSpec{},
	}

	docSpec := &rundown.DocumentSpec{
		ShortCodes: []*rundown.ShortCodeSpec{commandSpec},
		Options:    map[string]*rundown.ShortCodeOptionSpec{},
	}

	maxPos := 0
	var globOpt *rundown.ShortCodeOption

	for k, opt := range commandCode.Options {
		if opt.IsPositional {
			if opt.Position == -1 {
				globOpt = opt
				continue
			}

			if opt.Position > maxPos {
				maxPos = opt.Position
			}

			if len(args) > opt.Position && opt.Position >= 0 {
				commandSpec.Options[k] = &rundown.ShortCodeOptionSpec{Code: k, Value: args[opt.Position]}
			} else {
				return nil, fmt.Errorf("expected [%s] in position %d", strings.ToUpper(k), opt.Position)
			}
		} else if f := cmd.Flag(k); f != nil {
			commandSpec.Options[k] = &rundown.ShortCodeOptionSpec{Code: k, Value: f.Value.String()}
		}
	}

	if len(args) > maxPos {
		globSpec := &rundown.ShortCodeOptionSpec{Code: globOpt.Code, Value: ""}

		for _, arg := range args[maxPos+1:] {
			globSpec.Value = fmt.Sprintf("%s \"%s\"", globSpec.Value, arg)
		}

		globSpec.Value = strings.Trim(globSpec.Value, " ")
		commandSpec.Options[globSpec.Code] = globSpec
	}

	for k, _ := range docShortCodes.Options {
		if f := cmd.Flag(k); f != nil {
			docSpec.Options[k] = &rundown.ShortCodeOptionSpec{Code: k, Value: f.Value.String()}
		}
	}

	// fmt.Printf("DocSpec generated %s\n", docSpec)

	return docSpec, nil
}

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
