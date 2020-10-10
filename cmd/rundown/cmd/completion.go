package cmd

import (
	"fmt"
	"os"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(yourprogram completion bash)

# To load completions for each session, execute once:
Linux:
  $ yourprogram completion bash > /etc/bash_completion.d/yourprogram
MacOS:
  $ yourprogram completion bash > /usr/local/etc/bash_completion.d/yourprogram

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ yourprogram completion zsh > "${fpath[1]}/_yourprogram"

# You will need to start a new shell for this setup to take effect.

Fish:

$ yourprogram completion fish | source

# To load completions for each session, execute once:
$ yourprogram completion fish > ~/.config/fish/completions/yourprogram.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func performCompletion(args []string) []string {
	var result = []string{}

	if args[0] == argFilename {
		args = args[1:]
	}

	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	shortCodes := rd.GetShortCodes()
	if len(shortCodes.Codes) == 0 && len(shortCodes.Options) == 0 {
		return []string{}
	}

	if canSpecifyDocOpts(args) {
		// Show document codes and shortcodes
		for opt, optData := range shortCodes.Options {
			result = append(result, fmt.Sprintf("+%s=\t[%s] %s", opt, optData.Type, optData.Description))
		}

		for codeName, code := range shortCodes.Codes {
			if codeName == "rundown:help" {
				continue
			}

			result = append(result, fmt.Sprintf("%s\t%s", codeName, code.Name))
		}

		return result
	}

	if current := currentShortcode(args); current != "" {
		// Show code options
		currentCode := shortCodes.Codes[current]

		if currentCode != nil {
			for opt, optData := range currentCode.Options {
				result = append(result, fmt.Sprintf("+%s=\t[%s] %s", opt, optData.Type, optData.Description))
			}
		}
	}

	for codeName, code := range shortCodes.Codes {
		if codeName == "rundown:help" {
			continue
		}

		result = append(result, fmt.Sprintf("%s\t%s", codeName, code.Name))
	}

	return result
}

func currentShortcode(args []string) string {
	current := ""

	for _, a := range args {
		if a[0] != '+' {
			current = a
		}
	}

	return current
}

func canSpecifyDocOpts(args []string) bool {
	for _, a := range args {
		if a[0] != '+' {
			return false
		}
	}

	return true
}
