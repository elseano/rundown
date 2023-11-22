package exec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeCommentsToSpinnerCommands(t *testing.T) {
	result := ChangeCommentsToSpinnerCommands("bash", []byte("#> Do something\nrun_me"))

	require.Equal(t, "echo -n -e \"\x1b]R;SETSPINNER Do something\x9c\"\nrun_me", string(result))

	result = ChangeCommentsToSpinnerCommands("bash", []byte("if true; then\n  #> Do something\n  run_me\n fi"))
	require.Equal(t, "if true; then\n  echo -n -e \"\x1b]R;SETSPINNER Do something\x9c\"\n  run_me\n fi", string(result))
}
