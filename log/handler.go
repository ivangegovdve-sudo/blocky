package log

import (
	"context"
	"log/slog"
)

// contextHandler injects request-scoped attrs stored in the context into each
// emitted record. Because injection happens in Handle (after the level check),
// any slog.LogValuer attrs are resolved only when a record is actually emitted.
type contextHandler struct {
	next slog.Handler
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := attrsFromCtx(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name)}
}

// indentHandler prepends a fixed indent string to every record message. Used
// for the startup config dump; not on the hot path.
type indentHandler struct {
	indent string
	next   slog.Handler
}

func (h *indentHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *indentHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Message = h.indent + r.Message

	return h.next.Handle(ctx, r)
}

func (h *indentHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &indentHandler{indent: h.indent, next: h.next.WithAttrs(attrs)}
}

func (h *indentHandler) WithGroup(name string) slog.Handler {
	return &indentHandler{indent: h.indent, next: h.next.WithGroup(name)}
}
