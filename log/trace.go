package log

import (
	"context"
	"log/slog"
)

// Trace logs at the custom trace level (below Debug). args follow slog's
// (msg, args...) convention; pass slog.Attr or alternating key/value pairs.
func Trace(ctx context.Context, l *slog.Logger, msg string, args ...any) {
	l.Log(ctx, LevelTrace, msg, args...)
}
