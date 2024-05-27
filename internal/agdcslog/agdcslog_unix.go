//go:build !windows

package agdcslog

import (
	"log/syslog"
)

// systemLogger is a wrapper around [eventlog.Log].
type systemLogger struct {
	writer *syslog.Writer
}

// newSystemLogger returns a unix specific system logger.
func newSystemLogger(tag string) (l Logger, err error) {
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
var _ Logger = (*systemLogger)(nil)

// Debug implements [Logger] interface for *systemLogger.
func (l *systemLogger) Debug(msg string) (err error) {
	return l.writer.Debug(msg)
}

// Info implements [Logger] interface for *systemLogger.
func (l *systemLogger) Info(msg string) (err error) {
	return l.writer.Info(msg)
}

// Warning implements [Logger] interface for *systemLogger.
func (l *systemLogger) Warning(msg string) (err error) {
	return l.writer.Warning(msg)
}

// Error implements [Logger] interface for *systemLogger.
func (l *systemLogger) Error(msg string) (err error) {
	return l.writer.Err(msg)
}

// Close implements [Logger] interface for *systemLogger.
func (l *systemLogger) Close() (err error) {
	return l.writer.Close()
}
