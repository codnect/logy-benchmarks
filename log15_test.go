package benchmarks

import (
	"io"

	"gopkg.in/inconshreveable/log15.v2"
)

func newLog15() log15.Logger {
	logger := log15.New()
	logger.SetHandler(log15.StreamHandler(io.Discard, log15.JsonFormat()))
	return logger
}
