package util

import (
	"fmt"
	"sort"
	"strings"
)

// Replaces appearances of environment variables in source with variables present in the given environment.
// Replaces from longest environment variable name first.
func SubEnv(environment map[string]string, source string) string {
	keys := make([]string, 0, len(environment))
	for k := range environment {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) < len(keys[j]) })

	for _, k := range keys {
		source = strings.ReplaceAll(source, fmt.Sprintf("$%s", k), environment[k])
		source = strings.ReplaceAll(source, fmt.Sprintf("${%s}", k), environment[k])
	}

	return source
}
