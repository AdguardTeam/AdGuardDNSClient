package agdcslog_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcslog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogger is a mock implementation of [agdcslog.SystemLogger] interface for
// tests.
//
// TODO(e.burkov):  Move into a separate package with testing utilities.
type testLogger struct {
	onDebug   func(msg string) (err error)
	onInfo    func(msg string) (err error)
	onWarning func(msg string) (err error)
	onError   func(msg string) (err error)
	onClose   func() (err error)
}

// newTestLogger returns a new mock logger with all its methods set to panic.
func newTestLogger() (l *testLogger) {
	return &testLogger{
		onInfo: func(msg string) (_ error) {
			panic(fmt.Errorf("unexpected call to Info(%q)", msg))
		},
		onWarning: func(msg string) (_ error) {
			panic(fmt.Errorf("unexpected call to Warning(%q)", msg))
		},
		onError: func(msg string) (_ error) {
			panic(fmt.Errorf("unexpected call to Error(%q)", msg))
		},
		onDebug: func(msg string) (_ error) {
			panic(fmt.Errorf("unexpected call to Debug(%q)", msg))
		},
		onClose: func() (_ error) {
			panic(fmt.Errorf("unexpected call to Close"))
		},
	}
}

// type check
var _ agdcslog.SystemLogger = (*testLogger)(nil)

// Debug implements [agdcslog.SystemLogger] interface for *testLogger.
func (l *testLogger) Debug(msg string) (err error) {
	return l.onDebug(msg)
}

// Info implements [agdcslog.SystemLogger] interface for *testLogger.
func (l *testLogger) Info(msg string) (err error) {
	return l.onInfo(msg)
}

// Warning implements [agdcslog.SystemLogger] interface for *testLogger.
func (l *testLogger) Warning(msg string) (err error) {
	return l.onWarning(msg)
}

// Error implements [agdcslog.SystemLogger] interface for *testLogger.
func (l *testLogger) Error(msg string) (err error) {
	return l.onError(msg)
}

// Close implements [agdcslog.SystemLogger] interface for *testLogger.
func (l *testLogger) Close() (err error) {
	return l.onClose()
}

func TestSyslogHandler_Handle(t *testing.T) {
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

	l := newTestLogger()
	l.onInfo = outputWrite
	l.onWarning = outputWrite
	l.onError = outputWrite
	l.onDebug = outputWrite

	handler := agdcslog.NewSyslogHandler(l, &slog.HandlerOptions{
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

func TestSyslogHandler_Handle_race(t *testing.T) {
	var (
		mu     = sync.Mutex{}
		output = &bytes.Buffer{}
	)

	l := newTestLogger()
	l.onInfo = func(msg string) (err error) {
		mu.Lock()
		defer mu.Unlock()

		output.WriteString(msg + "\n")

		return nil
	}

	h := agdcslog.NewSyslogHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(h)

	// Test with multiple goroutines to be sure there are no races.
	const numGoroutine = 1_000

	wg := &sync.WaitGroup{}
	for range numGoroutine {
		wg.Add(1)

		go func() {
			defer wg.Done()

			logger.Info("test message", "attr", "abc")
		}()
	}

	wg.Wait()

	const wantMsg = `level=INFO msg="test message" attr=abc` + "\n"

	var num int
	for s := range strings.Lines(output.String()) {
		assert.Equal(t, wantMsg, s)
		num++
	}

	assert.Equal(t, numGoroutine, num)
}

func BenchmarkSyslogHandler_Handle(b *testing.B) {
	l := newTestLogger()
	l.onInfo = func(_ string) (_ error) { return nil }

	h := agdcslog.NewSyslogHandler(l, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	r.AddAttrs(
		slog.Int("int", 123),
		slog.String("string", "abc"),
	)

	var err error
	b.ReportAllocs()
	for b.Loop() {
		err = h.Handle(ctx, r)
	}

	require.NoError(b, err)

	// Most recent results:
	//
	//	goos: darwin
	//	goarch: amd64
	//	pkg: github.com/AdguardTeam/AdGuardDNSClient/internal/agdcslog
	//	cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
	//	BenchmarkSyslogHandler_Handle-12    	 2537618	       471.6 ns/op	      64 B/op	       1 allocs/op
}
