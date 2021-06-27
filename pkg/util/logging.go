package util

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func init() {
	Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func RedirectLogger(w io.Writer) {
	Logger = log.Output(w)
}
