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

const (
	// KeyUpstreamType is the log attribute for the upstream types.  See the
	// UpstreamType* constants below.
	KeyUpstreamType = "upstream_type"

	// KeyUpstreamGroup is the log attribute for the upstream groups.
	KeyUpstreamGroup = "upstream_group"
)

const (
	// UpstreamTypeBootstrap is the log attribute value for bootstrap upstreams.
	UpstreamTypeBootstrap = "bootstrap"

	// UpstreamTypeFallback is the log attribute value for fallback upstreams.
	UpstreamTypeFallback = "fallback"

	// UpstreamTypeMain is the log attribute value for main upstreams.
	UpstreamTypeMain = "main"
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

// SyslogHandler is a [slog.Handler] that writes to system log.
type SyslogHandler struct {
	logger      SystemLogger
	level       slog.Leveler
	bufTextPool *syncutil.Pool[bufferedTextHandler]
	attrs       []slog.Attr
}

// NewSyslogHandler returns an initialized SyslogHandler that writes to system
// log.  opts must not be nil and contain Level.
func NewSyslogHandler(logger SystemLogger, opts *slog.HandlerOptions) (h *SyslogHandler) {
	return &SyslogHandler{
		logger: logger,
		level:  opts.Level,
		bufTextPool: syncutil.NewPool(func() (bufTextHdlr *bufferedTextHandler) {
			return newBufferedTextHandler(initLineLenEst, opts)
		}),
		attrs: nil,
	}
}

// type check
var _ slog.Handler = (*SyslogHandler)(nil)

// Enabled implements the [slog.Handler] interface for *SyslogHandler.
func (h *SyslogHandler) Enabled(_ context.Context, level slog.Level) (enabled bool) {
	return level >= h.level.Level()
}

// Handle implements the [slog.Handler] interface for *SyslogHandler.
func (h *SyslogHandler) Handle(ctx context.Context, rec slog.Record) (err error) {
	bufTextHdlr := h.bufTextPool.Get()
	defer h.bufTextPool.Put(bufTextHdlr)

	bufTextHdlr.reset()

	rec.AddAttrs(h.attrs...)

	// System log entries already have a timestamp, so setting
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

// WithAttrs implements the [slog.Handler] interface for *SyslogHandler.
func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) (handler slog.Handler) {
	return &SyslogHandler{
		logger:      h.logger,
		level:       h.level,
		bufTextPool: h.bufTextPool,
		attrs:       append(slices.Clip(h.attrs), attrs...),
	}
}

// WithGroup implements the [slog.Handler] interface for *SyslogHandler.
func (h *SyslogHandler) WithGroup(name string) (handler slog.Handler) {
	return h
}

// Close closes an underlying system logger.
func (h *SyslogHandler) Close() (err error) {
	return h.logger.Close()
}
