//go:build linux

package agdcslog

import (
	"context"
	"log/syslog"
)

// systemLogger is the implementation of the [SystemLogger] interface for Linux.
type systemLogger struct {
	writer *syslog.Writer
}

// newSystemLogger returns a Linux-specific system logger.
func newSystemLogger(_ context.Context, tag string) (l SystemLogger, err error) {
	const priority = syslog.LOG_NOTICE | syslog.LOG_USER

	w, err := syslog.New(priority, tag)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	return &systemLogger{
		writer: w,
	}, nil
}

// type check
var _ SystemLogger = (*systemLogger)(nil)

// Debug implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Debug(msg string) (err error) {
	return l.writer.Debug(msg)
}

// Info implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Info(msg string) (err error) {
	return l.writer.Info(msg)
}

// Warning implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Warning(msg string) (err error) {
	return l.writer.Warning(msg)
}

// Error implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Error(msg string) (err error) {
	return l.writer.Err(msg)
}

// Close implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Close() (err error) {
	return l.writer.Close()
}
