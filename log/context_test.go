package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestCtxWithFieldsAndFromCtx(t *testing.T) {
	var buf bytes.Buffer
	configureTo(&buf, &Config{Level: Level(slog.LevelInfo), Format: FormatTypeJson, Timestamp: true})

	ctx, l := CtxWithFields(context.Background(), slog.String("req_id", "r1"))
	l.InfoContext(ctx, "first")

	// A logger from the same ctx must also carry the field.
	FromCtx(ctx).InfoContext(ctx, "second")

	out := buf.String()
	if strings.Count(out, `"req_id":"r1"`) != 2 {
		t.Errorf("expected req_id on both lines, got: %s", out)
	}
}
