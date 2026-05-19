// Package logger provides a small injectable logging surface for the conversion pipeline.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Level is a log severity (higher values are more severe).
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns a stable lowercase name (e.g. for CLI flags and JSON).
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

// ParseLevel maps common names to [Level]. Empty string defaults to [LevelInfo].
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level %q (want debug|info|warn|error)", s)
	}
}

// Logger receives structured key-value pairs after msg (even count: key, value, …).
type Logger interface {
	Log(level Level, msg string, kv ...any)
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
}

var (
	mu       sync.RWMutex
	defaultL Logger = Nop{}
)

// SetDefault replaces the process-wide logger.
func SetDefault(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	if l == nil {
		defaultL = Nop{}
		return
	}
	defaultL = l
}

// Default returns the process-wide logger.
func Default() Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultL
}

// Resolve returns l when non-nil, otherwise [Default].
func Resolve(l Logger) Logger {
	if l != nil {
		return l
	}
	return Default()
}

// Nop discards all log lines.
type Nop struct{}

func (Nop) Log(Level, string, ...any) {}
func (Nop) Debug(string, ...any)     {}
func (Nop) Info(string, ...any)      {}
func (Nop) Warn(string, ...any)      {}
func (Nop) Error(string, ...any)     {}

// minLevel filters lines below min before forwarding to inner.
type minLevel struct {
	min   Level
	inner Logger
}

// WithMinLevel returns a logger that drops events below min.
func WithMinLevel(inner Logger, min Level) Logger {
	inner = Resolve(inner)
	return minLevel{min: min, inner: inner}
}

func (m minLevel) enabled(level Level) bool { return level >= m.min }

func (m minLevel) Log(level Level, msg string, kv ...any) {
	if m.enabled(level) {
		m.inner.Log(level, msg, kv...)
	}
}

func (m minLevel) Debug(msg string, kv ...any) { m.Log(LevelDebug, msg, kv...) }
func (m minLevel) Info(msg string, kv ...any)  { m.Log(LevelInfo, msg, kv...) }
func (m minLevel) Warn(msg string, kv ...any)  { m.Log(LevelWarn, msg, kv...) }
func (m minLevel) Error(msg string, kv ...any) { m.Log(LevelError, msg, kv...) }

// Slog wraps a [slog.Logger] as [Logger].
type Slog struct{ L *slog.Logger }

// Std returns JSON logs on stderr at info level and above.
func Std() Logger { return StdAt(LevelInfo) }

// StdAt returns JSON logs on stderr at min level and above.
func StdAt(min Level) Logger {
	return Slog{L: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel(min)}))}
}

func (s Slog) Log(level Level, msg string, kv ...any) {
	if s.L == nil {
		return
	}
	s.L.Log(context.Background(), slogLevel(level), msg, kv...)
}

func (s Slog) Debug(msg string, kv ...any) { s.Log(LevelDebug, msg, kv...) }
func (s Slog) Info(msg string, kv ...any)  { s.Log(LevelInfo, msg, kv...) }
func (s Slog) Warn(msg string, kv ...any)  { s.Log(LevelWarn, msg, kv...) }
func (s Slog) Error(msg string, kv ...any) { s.Log(LevelError, msg, kv...) }

func slogLevel(l Level) slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
