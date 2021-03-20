module github.com/elseano/rundown

go 1.15

// Custom fork of Glamour supports rendering custom markdown nodes.
replace github.com/charmbracelet/glamour => github.com/elseano/glamour v0.2.1-0.20201024230852-96f145011f20

require (
	// github.com/alecthomas/chroma v0.8.0
	github.com/alecthomas/chroma v0.7.3
	github.com/briandowns/spinner v1.11.1
	github.com/buger/goterm v0.0.0-20200322175922-2f3e71b85129 // indirect
	github.com/c-bata/go-prompt v0.2.6 // indirect
	github.com/charmbracelet/glamour v0.2.0
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b
	github.com/creack/pty v1.1.11
	github.com/eliukblau/pixterm v1.3.1
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-errors/errors v1.1.1
	github.com/goccy/go-yaml v1.8.2
	github.com/hpcloud/tail v1.0.0
	github.com/kr/pty v1.1.1
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mjibson/esc v0.2.0 // indirect
	github.com/muesli/reflow v0.1.0
	github.com/muesli/termenv v0.7.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/papertrail/go-tail v0.0.0-20180509224916-973c153b0431
	github.com/paulrademacher/climenu v0.0.0-20151110221007-a1afbb4e378b
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tcnksm/go-input v0.0.0-20180404061846-548a7d7a8ee8 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/yuin/goldmark v1.2.1
	github.com/yuin/goldmark-highlighting v0.0.0-20200307114337-60d527fdb691
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.0.0-20210319071255-635bc2c9138d // indirect
	golang.org/x/tools v0.1.0
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
