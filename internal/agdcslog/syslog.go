package agdcslog

import (
	"context"
	"io"
)

// SystemLogger is a platform-specific system logger.
//
// TODO(e.burkov):  Consider moving to golibs.
//
// TODO(s.chzhen):  Add platform-specific tests.
type SystemLogger interface {
	// Debug logs a message at debug level.
	Debug(msg string) (err error)

	// Info logs a message at info level.
	Info(msg string) (err error)

	// Warning logs a message at warning level.
	Warning(msg string) (err error)

	// Error logs a message at error level.
	Error(msg string) (err error)

	// Close detaches from the system logger.
	io.Closer
}

// NewSystemLogger returns a platform-specific system logger.  name is the
// name of service.
func NewSystemLogger(ctx context.Context, name string) (l SystemLogger, err error) {
	return newSystemLogger(ctx, name)
}
