//go:build darwin

package agdcslog

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/ioutil"
	"github.com/c2h5oh/datasize"
)

// severity is a semantic alias for logging severity level name.
type severity = string

// Possible severity values.
const (
	sevDebug   severity = "debug"
	sevInfo    severity = "info"
	sevWarning severity = "warning"
	sevError   severity = "error"
)

// defaultOutLimit is the default limit of the output message size.  It's
// applied separately to stdout and stderr.
const defaultOutLimit datasize.ByteSize = 1 * datasize.KB

// systemLogger is the implementation of the [SystemLogger] interface for macOS,
// which uses the /usr/bin/logger command.
//
// TODO(e.burkov):  Get rid of it when golang/go#59229 is resolved.
type systemLogger struct {
	debug   *process
	info    *process
	warning *process
	error   *process

	// tag is the prefix for all log messages.
	//
	// TODO(e.burkov):  This is only needed because the /usr/bin/logger command
	// doesn't support the -t option.  Remove it if the situation changes.
	tag string
}

// newSystemLogger returns a macOS-specific system logger.
func newSystemLogger(tag string) (l SystemLogger, err error) {
	sysl := &systemLogger{
		tag: tag,
	}

	var errs []error

	// msgFmt is the local error message template.  It expects a severity as the
	// first argument and an error as the second.
	const msgFmt = "creating %s logger process: %w"

	sysl.debug, err = newProcess(sevDebug)
	if err != nil {
		errs = append(errs, fmt.Errorf(msgFmt, sevDebug, err))
	}

	sysl.info, err = newProcess(sevInfo)
	if err != nil {
		errs = append(errs, fmt.Errorf(msgFmt, sevInfo, err))
	}

	sysl.warning, err = newProcess(sevWarning)
	if err != nil {
		errs = append(errs, fmt.Errorf(msgFmt, sevWarning, err))
	}

	sysl.error, err = newProcess(sevError)
	if err != nil {
		errs = append(errs, fmt.Errorf(msgFmt, sevError, err))
	}

	if err = errors.Join(errs...); err != nil {
		return nil, errors.WithDeferred(err, sysl.Close())
	}

	return sysl, nil
}

// type check
var _ SystemLogger = (*systemLogger)(nil)

// Debug implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Debug(msg string) (err error) {
	// Don't wrap the error since it's informative enough as is.
	return l.debug.write(l.tag, msg)
}

// Info implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Info(msg string) (err error) {
	// Don't wrap the error since it's informative enough as is.
	return l.info.write(l.tag, msg)
}

// Warning implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Warning(msg string) (err error) {
	// Don't wrap the error since it's informative enough as is.
	return l.warning.write(l.tag, msg)
}

// Error implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Error(msg string) (err error) {
	// Don't wrap the error since it's informative enough as is.
	return l.error.write(l.tag, msg)
}

// Close implements the [SystemLogger] interface for *systemLogger.
func (l *systemLogger) Close() (err error) {
	defer func() { err = errors.Annotate(err, "closing logger processes: %w") }()

	var errs []error
	procs := []*process{
		l.debug,
		l.info,
		l.warning,
		l.error,
	}

	for _, p := range procs {
		if p == nil {
			continue
		}

		err = p.close()
		if err != nil {
			// Don't wrap the error since it's informative enough as is.
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// process is an instance of the logger process with a particular severity.
type process struct {
	// mu synchronizes writes to stdin between each other and with the process
	// closing.
	mu *sync.Mutex

	// cmd is the logger command of a particular severity.
	cmd *exec.Cmd

	// stdin is the pipe to stdin of cmd.
	stdin io.WriteCloser

	// stdout is the truncated standard output of cmd.
	stdout *bytes.Buffer

	// stderr is the truncated standard error of cmd.
	stderr *bytes.Buffer

	// severity is the severity level this logger process correponds to.  See
	// [severity].
	severity severity
}

// newProcess creates a new process with a particular severity.
func newProcess(sev severity) (p *process, err error) {
	const (
		binPath        = "/usr/bin/logger"
		optionPriority = "-p"
		facilityVal    = "user"
	)

	priorityVal := facilityVal + "." + sev

	// #nosec G204 -- Trust the variable to be a valid syslog priority since it
	// always constructed from the predefined constants.
	cmd := exec.Command(binPath, optionPriority, priorityVal)
	if cmd.Err != nil {
		return nil, fmt.Errorf("creating command: %w", cmd.Err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdin pipe: %w", err)
	}

	limit := uint(defaultOutLimit.Bytes())

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = ioutil.NewTruncatedWriter(stdout, limit)
	cmd.Stderr = ioutil.NewTruncatedWriter(stderr, limit)

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting command: %w", err)
	}

	return &process{
		mu:       &sync.Mutex{},
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		severity: sev,
	}, nil
}

// write writes the message to the logger.
func (p *process) write(tag, msg string) (err error) {
	defer func() { err = errors.Annotate(err, "writing %s message", p.severity) }()

	msg = strings.TrimSuffix(msg, "\n")

	p.mu.Lock()
	defer p.mu.Unlock()

	_, err = fmt.Fprintf(p.stdin, "%s: %s\n", tag, msg)

	return err
}

// close closes the process' pipes and waits for the command to exit.
func (p *process) close() (err error) {
	defer func() { err = errors.Annotate(err, "closing %s logger: %w", p.severity) }()

	var errs []error

	p.mu.Lock()
	defer p.mu.Unlock()

	if err = p.stdin.Close(); err != nil {
		err = fmt.Errorf("closing stdin: %w", err)
		errs = append(errs, err)
	}

	if err = p.cmd.Wait(); err != nil {
		err = fmt.Errorf("waiting: %w", err)
		errs = append(errs, err)
	}

	if p.stdout.Len() > 0 {
		err = fmt.Errorf("unexpected stdout output: %q", p.stdout)
		errs = append(errs, err)
	}

	if p.stderr.Len() > 0 {
		err = fmt.Errorf("unexpected stderr output: %q", p.stderr)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
