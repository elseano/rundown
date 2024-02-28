package term

import (
	"os"
)

type CI string

var GitHubCI CI = "github"
var GitLabCI CI = "gitlab"
var UnknownCI CI = "unknown"
var NoCI CI = "none"

func (c CI) IsCI() bool {
	return c != NoCI
}

func GetCI() CI {
	if _, gitlab := os.LookupEnv("GITLAB_CI"); gitlab {
		return GitLabCI
	} else if _, github := os.LookupEnv("GITHUB_ACTIONS"); github {
		return GitHubCI
	} else if _, ci := os.LookupEnv("CI"); ci {
		return UnknownCI
	}

	return NoCI
}
