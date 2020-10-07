package cmd

import (
	"errors"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/manifoldco/promptui"
)

type shortcodeDisplayItem struct {
	Name        string
	Code        string
	Description string
	Indent      string
}

func AskShortCode() (*rundown.ShortCodeSpec, error) {

	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	shortCodes := rd.GetShortCodes()
	if len(shortCodes.Codes) == 0 {
		return nil, errors.New("No shortcodes in document")
	}

	items := []shortcodeDisplayItem{
		shortcodeDisplayItem{
			Code:        "exit",
			Name:        "Exit",
			Description: "Exit & Quit Rundown",
		},
	}

	for _, shortCode := range shortCodes.Order {
		code := shortCodes.Codes[shortCode]

		indent := strings.Repeat("  ", code.Section.Level-1)
		if code.Section.Level > 1 {
			indent = indent + "- "
		}

		item := shortcodeDisplayItem{
			Code:        code.Code,
			Name:        code.Name,
			Description: code.Description,
			Indent:      indent,
		}
		items = append(items, item)
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . | bold }}",
		Active:   "{{ \">\" | yellow }} {{ .Indent | faint }}{{ .Name | bold | cyan }} ({{ .Code }})",
		Inactive: "  {{ .Indent | faint }}{{ .Name | cyan }} ({{ .Code | faint }})",
		Selected: "{{ \">\" | blue }} {{ .Name | blue }}",
		Details:  "\n{{ printf \"%." + strconv.Itoa(util.GetConsoleWidth()) + "s\" .Description | faint }}",
	}

	searcher := func(input string, index int) bool {
		item := items[index]
		name := strings.Replace(strings.ToLower(item.Name), " ", "", -1)
		code := strings.Replace(strings.ToLower(item.Code), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input) || strings.Contains(code, input)
	}

	stdin := exec.NewStdinReader()

	prompt := promptui.Select{
		Label:     "What to run",
		Items:     items,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
		Stdin:     stdin.Claim(),
	}

	i, _, err := prompt.Run()

	stdin.Stop()

	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, nil
	}

	spec := &rundown.ShortCodeSpec{
		Code:    items[i].Code,
		Options: map[string]*rundown.ShortCodeOptionSpec{},
	}

	return spec, nil
}
