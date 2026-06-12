package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestTraceEmitsAtTraceLevel(t *testing.T) {
	var buf bytes.Buffer
	configureTo(&buf, &Config{Level: Level(LevelTrace), Format: FormatTypeJson, Timestamp: false})

	Trace(context.Background(), Log(), "tracing", slog.String("k", "v"))

	out := buf.String()
	if !strings.Contains(out, `"tracing"`) || !strings.Contains(out, `"k":"v"`) {
		t.Errorf("expected trace line with attr, got: %s", out)
	}
	if !strings.Contains(out, "TRACE") {
		t.Errorf("expected TRACE level label, got: %s", out)
	}
}

func TestTraceSilentBelowTrace(t *testing.T) {
	var buf bytes.Buffer
	configureTo(&buf, &Config{Level: Level(slog.LevelDebug), Format: FormatTypeJson, Timestamp: false})

	Trace(context.Background(), Log(), "should be hidden")

	if buf.Len() != 0 {
		t.Errorf("expected no output below trace level, got: %s", buf.String())
	}
}
