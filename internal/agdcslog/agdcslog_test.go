package agdcslog_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcslog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLogger is a mock implementation of [agdcslog.Logger] interface for tests.
type mockLogger struct {
	onDebug   func(msg string) (err error)
	onInfo    func(msg string) (err error)
	onWarning func(msg string) (err error)
	onError   func(msg string) (err error)
	onClose   func() (err error)
}

// NewMockLogger returns a new mock logger with placeholder methods.
func NewMockLogger() (l *mockLogger) {
	const errMsg = "not implemented"

	notImplemented := func(_ string) (_ error) {
		panic(errMsg)
	}

	return &mockLogger{
		onInfo:    notImplemented,
		onWarning: notImplemented,
		onError:   notImplemented,
		onDebug:   notImplemented,
		onClose: func() (_ error) {
			panic(errMsg)
		},
	}
}

// type check
var _ agdcslog.Logger = (*mockLogger)(nil)

// Debug implements [agdcslog.Logger] interface for *mockLogger.
func (l *mockLogger) Debug(msg string) (err error) {
	return l.onDebug(msg)
}

// Info implements [agdcslog.Logger] interface for *mockLogger.
func (l *mockLogger) Info(msg string) (err error) {
	return l.onInfo(msg)
}

// Warning implements [agdcslog.Logger] interface for *mockLogger.
func (l *mockLogger) Warning(msg string) (err error) {
	return l.onWarning(msg)
}

// Error implements [agdcslog.Logger] interface for *mockLogger.
func (l *mockLogger) Error(msg string) (err error) {
	return l.onError(msg)
}

// Close implements [agdcslog.Logger] interface for *mockLogger.
func (l *mockLogger) Close() (err error) {
	return l.onClose()
}

func TestSystemHandler_Handle(t *testing.T) {
	var (
		mu     = sync.Mutex{}
		output = &bytes.Buffer{}
	)

	outputWrite := func(msg string) (err error) {
		mu.Lock()
		defer mu.Unlock()

		output.WriteString(msg + "\n")

		return nil
	}

	l := NewMockLogger()
	l.onInfo = outputWrite
	l.onWarning = outputWrite
	l.onError = outputWrite
	l.onDebug = outputWrite

	handler := agdcslog.NewSystemHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	const testMsg = "test message"

	testCases := []struct {
		name  string
		want  string
		attrs []slog.Attr
		level slog.Level
	}{{
		name:  "level_info",
		level: slog.LevelInfo,
		want:  `level=INFO msg="test message"`,
	}, {
		name:  "level_warn",
		level: slog.LevelWarn,
		attrs: []slog.Attr{},
		want:  `level=WARN msg="test message"`,
	}, {
		name:  "level_err",
		level: slog.LevelError,
		attrs: []slog.Attr{},
		want:  `level=ERROR msg="test message"`,
	}, {
		name:  "level_debug",
		level: slog.LevelDebug,
		attrs: []slog.Attr{},
		want:  `level=DEBUG msg="test message"`,
	}, {
		name:  "level_custom",
		level: slog.Level(-8),
		attrs: []slog.Attr{},
		want:  `level=DEBUG-4 msg="test message"`,
	}, {
		name:  "level_info_with_args",
		level: slog.LevelInfo,
		attrs: []slog.Attr{
			slog.Int("int", 123),
			slog.String("string", "abc"),
		},
		want: `level=INFO msg="test message" int=123 string=abc`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := slog.NewRecord(time.Now(), tc.level, testMsg, 0)
			r.AddAttrs(tc.attrs...)

			err := handler.Handle(context.Background(), r)
			require.NoError(t, err)

			line, err := output.ReadString('\n')
			require.NoError(t, err)

			line = line[:len(line)-1]
			assert.Equal(t, tc.want, line)
		})
	}
}

func TestSystemHandler_Handle_race(t *testing.T) {
	var (
		mu     = sync.Mutex{}
		output = &bytes.Buffer{}
	)

	l := NewMockLogger()
	l.onInfo = func(msg string) (err error) {
		mu.Lock()
		defer mu.Unlock()

		output.WriteString(msg + "\n")

		return nil
	}

	h := agdcslog.NewSystemHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(h)

	// Test with multiple goroutines to be sure there are no races.
	const numGoroutine = 1_000

	wg := &sync.WaitGroup{}
	for i := 0; i < numGoroutine; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			logger.Info("test message", "attr", "abc")
		}(i)
	}

	wg.Wait()

	textOutputStrings := strings.Split(output.String(), "\n")

	const wantMsg = `level=INFO msg="test message" attr=abc`

	for i := 0; i < numGoroutine; i++ {
		assert.Equal(t, wantMsg, textOutputStrings[i])
	}
}

// errSink is a sink for benchmark results.
var errSink error

func BenchmarkSystemHandler_Handle(b *testing.B) {
	l := NewMockLogger()
	l.onInfo = func(_ string) (_ error) { return nil }

	h := agdcslog.NewSystemHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	r.AddAttrs(
		slog.Int("int", 123),
		slog.String("string", "abc"),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errSink = h.Handle(ctx, r)
	}

	require.NoError(b, errSink)

	// Most recent result, on a ThinkPad P15s with a Intel Core i7-10510U CPU:
	//	goos: linux
	//	goarch: amd64
	//	pkg: github.com/AdguardTeam/AdGuardDNSClient/internal/agdcslog
	//	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	//	BenchmarkSystemHandler_Handle-8   	 2595381	       448.2 ns/op	      64 B/op	       1 allocs/op
}
