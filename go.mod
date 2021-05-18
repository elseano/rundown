module github.com/elseano/rundown

go 1.15

// Custom fork of Glamour supports rendering custom markdown nodes.
replace github.com/charmbracelet/glamour => github.com/elseano/glamour v0.2.1-0.20201024230852-96f145011f20

require (
	// github.com/alecthomas/chroma v0.8.0
	github.com/alecthomas/chroma v0.7.3
	github.com/charmbracelet/glamour v0.2.0
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/eliukblau/pixterm v1.3.1
	github.com/fatih/color v1.10.0
	github.com/goccy/go-yaml v1.8.9
	github.com/kr/pty v1.1.1
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/muesli/termenv v0.7.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/rs/zerolog v1.22.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.4.0
	github.com/thecodeteam/goodbye v0.0.0-20170927022442-a83968bda2d3 // indirect
	github.com/yuin/goldmark v1.2.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4 // indirect
)
