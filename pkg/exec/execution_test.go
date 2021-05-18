package exec

import (
	"testing"

	"github.com/elseano/rundown/pkg/util"
	"github.com/elseano/rundown/testutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestExecutionBasic(t *testing.T) {
	script := []byte(`echo "Hello there"`)

	intent, err := NewExecution("/bin/bash", script)
	util.Logger = log.Output(zerolog.ConsoleWriter{Out: testutil.NewTestWriter(t)})

	if assert.NoError(t, err) {
		assert.NotNil(t, intent)

		result, err := intent.Execute()

		if assert.Nil(t, err) {
			assert.Equal(t, "Hello there\n", string(result.Output))
		}
	}
}

func TestExecutionFailure(t *testing.T) {
	script := []byte(`some_invalid_command`)

	intent, err := NewExecution("/bin/bash", script)
	util.Logger = log.Output(zerolog.ConsoleWriter{Out: testutil.NewTestWriter(t)})

	if assert.NoError(t, err) {
		assert.NotNil(t, intent)

		result, err := intent.Execute()

		assert.NoError(t, err)
		assert.Equal(t, 127, result.ExitCode)
		assert.Equal(t, "SCRIPT: line 3: some_invalid_command: command not found\n", string(result.Output))
	}
}

// func TestExecutionShebangFailure(t *testing.T) {
// 	script := []byte(``)

// 	intent, err := NewExecution("/bin/blahgggg", script)
// 	intent.Logger = log.Output(zerolog.ConsoleWriter{Out: testutil.NewTestWriter(t)})

// 	assert.Nil(t, err)
// 	assert.NotNil(t, intent)

// 	_, err = intent.Execute()

// 	if assert.Error(t, err) {
// 		assert.IsType(t, &InterpreterNotFound{}, err)
// 	}
// }

// func TestExecutionRPC(t *testing.T) {
// 	script := []byte(`env;echo SOMETHING >> $RDRPC`)
// 	var buffer strings.Builder

// 	intent, err := NewExecution("bash", script)
// 	util.Logger = log.Output(zerolog.ConsoleWriter{Out: testutil.NewTestWriter(t)})

// 	// intent.Subscribe(func(message []byte) {
// 	// 	buffer.Write(message)
// 	// 	buffer.Write([]byte(";"))
// 	// })

// 	if assert.NoError(t, err) {
// 		assert.NotNil(t, intent)

// 		_, err = intent.Execute()

// 		if assert.NoError(t, err) {
// 			assert.Contains(t, buffer.String(), ";SOMETHING;")
// 		}
// 	}

// }

func TestExecutionEnvironmentCapture(t *testing.T) {
	script := []byte(`export NEW_VALUE=TRUE`)

	intent, _ := NewExecution("bash", script)
	util.Logger = log.Output(zerolog.ConsoleWriter{Out: testutil.NewTestWriter(t)})

	result, err := intent.Execute()

	if assert.NoError(t, err) {
		assert.Equal(t, "TRUE", result.Env["NEW_VALUE"])
		assert.Len(t, result.Env, 1) // Should only include the new environment value.
	}

}
