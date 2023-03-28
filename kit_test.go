package benchmarks

import (
	"io"

	"github.com/go-kit/log"
)

func newKitLog(fields ...interface{}) log.Logger {
	return log.With(log.NewJSONLogger(io.Discard), fields...)
}
