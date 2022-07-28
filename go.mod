module github.com/elseano/rundown

go 1.15

replace github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 => github.com/elseano/go-ansiterm v0.0.0-20220406061920-d6b43c07a79c

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
	// github.com/alecthomas/chroma v0.8.0
	github.com/alecthomas/chroma v0.9.2
	github.com/charmbracelet/glamour v0.3.0
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/eliukblau/pixterm v1.3.1
	github.com/fatih/color v1.10.0
	github.com/go-playground/validator/v10 v10.9.0 // indirect
	github.com/goccy/go-yaml v1.8.9
	github.com/kr/pty v1.1.1
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.14
	github.com/microcosm-cc/bluemonday v1.0.14 // indirect
	github.com/muesli/reflow v0.3.0
	github.com/muesli/termenv v0.9.0
	github.com/rs/zerolog v1.22.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.7.0
	github.com/thecodeteam/goodbye v0.0.0-20170927022442-a83968bda2d3
	github.com/yuin/goldmark v1.4.12
	github.com/yuin/goldmark-emoji v1.0.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/sys v0.0.0-20211003122950-b1ebd4e1001c // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/guregu/null.v4 v4.0.0
)
