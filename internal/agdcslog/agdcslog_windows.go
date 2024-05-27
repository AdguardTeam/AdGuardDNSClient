//go:build windows

package agdcslog

import (
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

// Event IDs are the application-specific identifiers for the events.  They
// provide a unique identification of a particular event.
//
// See
// https://learn.microsoft.com/en-us/windows/win32/eventlog/event-identifiers
// and also
// https://github.com/kardianos/service/blob/master/service_windows.go.
const (
	infoEventID    = 1
	warningEventID = 2
	errorEventID   = 3
	debugEventID   = 4
)

// systemLogger is a wrapper around [eventlog.Log].
type systemLogger struct {
	writer *eventlog.Log
}

// newSystemLogger returns a windows specific system logger.
//
// Note that the eventlog src is the same as the service name.  Otherwise, we
// will get "the description for event id cannot be found" warning in every log
// record.
func newSystemLogger(src string) (l Logger, err error) {
	const events = eventlog.Info | eventlog.Warning | eventlog.Error

	// Continue if we receive "registry key already exists" or if we get
	// ERROR_ACCESS_DENIED so that we can log without administrative permissions
	// for pre-existing eventlog sources.
	err = eventlog.InstallAsEventCreate(src, events)
	if err != nil {
		if !strings.Contains(err.Error(), "exists") && err != windows.ERROR_ACCESS_DENIED {
			// Don't wrap the error since it's informative enough as is.
			return nil, err
		}
	}

	writer, err := eventlog.Open(src)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	}

	return &systemLogger{
		writer: writer,
	}, nil
}

// type check
var _ Logger = (*systemLogger)(nil)

// Debug implements [Logger] interface for *systemLogger.
func (l *systemLogger) Debug(msg string) (err error) {
	// Event Log doesn't have Debug log level, use Info.
	return l.writer.Info(debugEventID, msg)
}

// Info implements [Logger] interface for *systemLogger.
func (l *systemLogger) Info(msg string) (err error) {
	return l.writer.Info(infoEventID, msg)
}

// Warning implements [Logger] interface for *systemLogger.
func (l *systemLogger) Warning(msg string) (err error) {
	return l.writer.Warning(warningEventID, msg)
}

// Error implements [Logger] interface for *systemLogger.
func (l *systemLogger) Error(msg string) (err error) {
	return l.writer.Error(errorEventID, msg)
}

// Close implements [Logger] interface for *systemLogger.
func (l *systemLogger) Close() (err error) {
	return l.writer.Close()
}
