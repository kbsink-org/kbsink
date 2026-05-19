package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestResolve_nilUsesDefault(t *testing.T) {
	SetDefault(Nop{})
	got := Resolve(nil)
	if _, ok := got.(Nop); !ok {
		t.Fatalf("expected Nop, got %T", got)
	}
}

func TestWithMinLevel_filtersDebug(t *testing.T) {
	var buf bytes.Buffer
	inner := Slog{L: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	log := WithMinLevel(inner, LevelInfo)
	log.Debug("hidden")
	log.Info("visible")
	if strings.Contains(buf.String(), "hidden") {
		t.Fatalf("debug should be filtered: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "visible") {
		t.Fatalf("info should appear: %s", buf.String())
	}
}

func TestParseLevel(t *testing.T) {
	l, err := ParseLevel("warn")
	if err != nil || l != LevelWarn {
		t.Fatalf("got %v %v", l, err)
	}
	_, err = ParseLevel("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}
