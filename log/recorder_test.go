package log

import (
	"context"
	"log/slog"
	"testing"
)

func TestRecorderCapturesMessages(t *testing.T) {
	logger, rec := NewRecorder()

	logger.Info("first")
	logger.Warn("second", slog.String("k", "v"))

	if got := rec.Messages(); len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("Messages() = %v", got)
	}
	if rec.LastMessage() != "second" {
		t.Errorf("LastMessage() = %q", rec.LastMessage())
	}

	rec.Reset()
	if len(rec.Messages()) != 0 {
		t.Error("Reset did not clear messages")
	}
}

func TestCaptureGlobalRestores(t *testing.T) {
	before := Log()

	rec, restore := CaptureGlobal()
	Log().InfoContext(context.Background(), "captured")
	if rec.LastMessage() != "captured" {
		t.Errorf("expected captured message, got %q", rec.LastMessage())
	}

	restore()
	if Log() != before {
		t.Error("CaptureGlobal restore did not reinstate the previous logger")
	}
}
