// The styles here are copied from the awesome Glamour library
package styles

type Style struct {
	Text                string
	Error               string
	Comment             string
	CommentPreproc      string
	Keyword             string
	KeywordReserved     string
	KeywordNamespace    string
	KeywordType         string
	Operator            string
	Punctuation         string
	Name                string
	NameBuiltin         string
	NameTag             string
	NameAttribute       string
	NameClass           string
	NameConstant        string
	NameDecorator       string
	NameException       string
	NameFunction        string
	NameOther           string
	Literal             string
	LiteralNumber       string
	LiteralDate         string
	LiteralString       string
	LiteralStringEscape string
	GenericDeleted      string
	GenericEmph         string
	GenericInserted     string
	GenericStrong       string
	GenericSubheading   string
}

var Dark Style = Style{
	Text:                "#c4c4c4",
	Error:               "#f1f1f1 bg:#f05b5b",
	Comment:             "#676767",
	CommentPreproc:      "#FF875F",
	Keyword:             "#00AAFF",
	KeywordReserved:     "#FF5FD2",
	KeywordNamespace:    "#FF5F87",
	KeywordType:         "#6E6ED8",
	Operator:            "#EF8080",
	Punctuation:         "#E8E8A8",
	Name:                "#c4c4c4",
	NameBuiltin:         "#FF8EC7",
	NameTag:             "#B083EA",
	NameAttribute:       "#7A7AE6",
	NameClass:           "#F1F1F1 underline bold",
	NameDecorator:       "#FFFF87",
	NameFunction:        "#00D787",
	LiteralNumber:       "#6EEFC0",
	LiteralString:       "#C69669",
	LiteralStringEscape: "#AFFFD7",
	GenericDeleted:      "#FD5B5B",
	GenericEmph:         "italic",
	GenericInserted:     "#00D787",
	GenericStrong:       "bold",
	GenericSubheading:   "#777777",
}
