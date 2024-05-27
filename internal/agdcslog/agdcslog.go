// Package agdcslog contains slog handler implementation that writes to system
// log.
package agdcslog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/AdguardTeam/golibs/syncutil"
)

// initLineLenEst is the estimation used to set the initial sizes of log-line
// buffers.
const initLineLenEst = 256

// bufferedTextHandler is a combination of bytes buffer and a text handler that
// writes to it.
type bufferedTextHandler struct {
	buffer  *bytes.Buffer
	handler *slog.TextHandler
}

// newBufferedTextHandler returns a new bufferedTextHandler with the given
// buffer length.
func newBufferedTextHandler(l int, opts *slog.HandlerOptions) (h *bufferedTextHandler) {
	buf := bytes.NewBuffer(make([]byte, 0, l))

	return &bufferedTextHandler{
		buffer:  buf,
		handler: slog.NewTextHandler(buf, opts),
	}
}

// reset must be called before using h after retrieving it from a pool.
func (h *bufferedTextHandler) reset() {
	h.buffer.Reset()
}

// Logger is a platform-specific system Logger.
type Logger interface {
	Debug(msg string) (err error)
	Info(msg string) (err error)
	Warning(msg string) (err error)
	Error(msg string) (err error)
	Close() (err error)
}

// NewSystemLogger returns a platform-specific system logger that writes to
// system log.  name is the service name.
func NewSystemLogger(name string) (l Logger, err error) {
	return newSystemLogger(name)
}

// SystemHandler is a [slog.Handler] that writes to system log.
type SystemHandler struct {
	logger      Logger
	level       slog.Leveler
	bufTextPool *syncutil.Pool[bufferedTextHandler]
	attrs       []slog.Attr
}

// NewSystemHandler returns an initialized SystemHandler that writes to system
// log.  opts must not be nil and contain Level.
func NewSystemHandler(logger Logger, opts *slog.HandlerOptions) (h *SystemHandler) {
	return &SystemHandler{
		logger: logger,
		level:  opts.Level,
		bufTextPool: syncutil.NewPool(func() (bufTextHdlr *bufferedTextHandler) {
			return newBufferedTextHandler(initLineLenEst, opts)
		}),
		attrs: nil,
	}
}

// type check
var _ slog.Handler = (*SystemHandler)(nil)

// Enabled implements the [slog.Handler] interface for *SystemHandler.
func (h *SystemHandler) Enabled(_ context.Context, level slog.Level) (enabled bool) {
	return level >= h.level.Level()
}

// Handle implements the [slog.Handler] interface for *SystemHandler.
func (h *SystemHandler) Handle(ctx context.Context, rec slog.Record) (err error) {
	bufTextHdlr := h.bufTextPool.Get()
	defer h.bufTextPool.Put(bufTextHdlr)

	bufTextHdlr.reset()

	rec.AddAttrs(h.attrs...)

	// System log entires already have a timestamp, so setting
	// [slog.Record.Time] to zero time will cause it to be ignored by the slog
	// text handler.
	//
	// TODO(s.chzhen):  Allow timestamp.
	rec.Time = time.Time{}

	err = bufTextHdlr.handler.Handle(ctx, rec)
	if err != nil {
		return fmt.Errorf("handling text for msg: %w", err)
	}

	msg := bufTextHdlr.buffer.String()

	// Remove newline.
	msg = msg[:len(msg)-1]

	switch rec.Level {
	case slog.LevelDebug:
		err = h.logger.Debug(msg)
	case slog.LevelInfo:
		err = h.logger.Info(msg)
	case slog.LevelWarn:
		err = h.logger.Warning(msg)
	case slog.LevelError:
		err = h.logger.Error(msg)
	default:
		err = h.logger.Info(msg)
	}

	return err
}

// WithAttrs implements the [slog.Handler] interface for *SystemHandler.
func (h *SystemHandler) WithAttrs(attrs []slog.Attr) (handler slog.Handler) {
	return &SystemHandler{
		logger:      h.logger,
		level:       h.level,
		bufTextPool: h.bufTextPool,
		attrs:       append(slices.Clip(h.attrs), attrs...),
	}
}

// WithGroup implements the [slog.Handler] interface for *SystemHandler.
func (h *SystemHandler) WithGroup(name string) (handler slog.Handler) {
	return h
}

// Close closes an underlying system logger.
func (h *SystemHandler) Close() (err error) {
	return h.logger.Close()
}
