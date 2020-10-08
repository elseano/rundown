package rundown

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/yuin/goldmark/ast"

	"github.com/elseano/rundown/pkg/markdown"
)

type rundownHandler struct {
	// markdown.RundownHandler
	ctx *Context
}

func NewRundownHandler(ctx *Context) *rundownHandler {
	return &rundownHandler{
		ctx: ctx,
	}
}

func (v *rundownHandler) Mutate(input []byte, node ast.Node) ([]byte, error) {
	rundown := node.(markdown.RundownNode)

	if rundown.GetModifiers().Flags[EnvAwareFlag] {
		result, err := SubEnv(string(input), v.ctx)

		if err != nil {
			return input, err
		}

		return []byte(result), nil
	}

	return input, nil
}

func (v *rundownHandler) OnRundownNode(node ast.Node, entering bool) error {
	if !entering {
		if rundown, ok := node.(*markdown.RundownBlock); ok {
			if rundown.GetModifiers().Flags[StopOkFlag] {
				return &StopError{Result: StopOkResult}
			}

			if rundown.GetModifiers().Flags[StopFailFlag] {
				return &StopError{Result: StopFailResult}
			}

			if setEnv, ok := rundown.GetModifiers().Flags[markdown.Flag("set-env")]; setEnv && ok {
				for k, val := range rundown.GetModifiers().Values {
					envName := strings.ReplaceAll(strings.ToUpper(string(k)), "-", "_")
					v.ctx.SetEnv(envName, val)
				}
			}

			if rundown.GetModifiers().HasAny("invoke") {
				source := rundown.GetModifiers().GetValue("from")

				if source == nil {
					source = &v.ctx.CurrentFile
				}

				name := rundown.GetModifiers().GetValue("invoke")
				if name == nil {
					return errors.New("Invoke requires a ShortCode value")
				}

				rd, err := LoadFile(*source)
				if err != nil {
					return err
				}

				if info := rd.GetShortCodes().Functions[*name]; info != nil {
					section := info.Section

					mods := markdown.NewModifiers()
					mods.Flags[markdown.Flag("set-env")] = true

					for k, v := range rundown.GetModifiers().Values {
						if strings.HasPrefix(string(k), "opt-") {
							mods.Values[k] = v
						}
					}

					// Create a rundown block which sets the environment to the invoke options.
					envNode := markdown.NewRundownBlock(mods)

					node.Parent().InsertAfter(node.Parent(), node, section)

					// Remove the heading when invoking functions, unless we specify we want to keep the heading
					if keepHeading, specified := mods.Flags[markdown.Flag("keep-heading")]; keepHeading == false || !specified {
						section.RemoveChild(section, section.FirstChild())
					}

					// Add the environment setting. FIXME - We should nest the Section inside this node to wrap the context.
					section.InsertBefore(section, section.FirstChild(), envNode)
				} else {
					// ShortCode not found in file.
					if flag, ok := rundown.GetModifiers().Flags["ignore-missing"]; flag && ok {
						return nil
					}

					return errors.New("Cannot find " + *name + " in " + *source)
				}

			}
		}
	}

	return nil
}

func (v *rundownHandler) OnExecute(node *markdown.ExecutionBlock, source []byte, out io.Writer) (markdown.ExecutionResult, error) {
	result := Execute(v.ctx, node, source, v.ctx.Logger, os.Stdout)
	switch result {
	case SuccessfulExecution:
		return markdown.Continue, nil
	case SkipToNextHeading:

		nextNode := node.NextSibling()

		deleteNodesUntilSection(nextNode)

		return markdown.Continue, nil
	case StopOkResult:
		return markdown.Stop, nil
	case StopFailResult:
		return markdown.Stop, errors.New("Stop requested")
	}

	return markdown.Stop, &StopError{Result: result}
}

func deleteNodesUntilSection(node ast.Node) {
	for node != nil {
		if _, isSection := node.(*markdown.Section); isSection {
			return
		}

		nextNode := node.NextSibling()
		if nextNode == nil {
			nextNode = node.Parent().NextSibling()
		}

		node.Parent().RemoveChild(node.Parent(), node)

		node = nextNode
	}
}
