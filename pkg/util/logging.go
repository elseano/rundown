package util

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func init() {
	Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
