package cli

import (
	"strings"

	"github.com/elseano/rundown/pkg/segments"
	"github.com/manifoldco/promptui"
)

type ShortcodeItem struct {
	Name        string
	ShortCode   string
	Description string
	Level       string
}

func processHeadings(doc []segments.Segment) []*ShortcodeItem {
	var headings = []*ShortcodeItem{}

	for _, seg := range doc {
		if heading, ok := seg.(*segments.HeadingMarker); ok {
			var indent = ""

			if heading.Level > 1 {
				indent = strings.Repeat("  ", heading.Level-2) + " - "
			}
			item := &ShortcodeItem{
				Name:        heading.Title,
				ShortCode:   heading.ShortCode,
				Level:       indent,
				Description: heading.Description,
			}

			headings = append(headings, item)
		}
	}

	return headings
}

func ShortcodeMenu(doc []segments.Segment) string {
	KillReadlineBell()

	peppers := processHeadings(doc)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . | bold }}",
		Active:   "{{ \">\" | yellow }} {{ .Level | faint }}{{ .Name | bold | cyan }} ({{ .ShortCode }})",
		Inactive: "  {{ .Level | faint }}{{ .Name | cyan }} ({{ .ShortCode | faint }})",
		Selected: "{{ \">\" | blue }} {{ .Name | blue }}",
		Details: `
     {{ .Description | faint }}`,
	}

	searcher := func(input string, index int) bool {
		pepper := peppers[index]
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "What to run",
		Items:     peppers,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()

	if err != nil {
		return ""
	}

	return peppers[i].ShortCode
}
