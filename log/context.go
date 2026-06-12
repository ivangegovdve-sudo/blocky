package log

import (
	"context"
	"log/slog"
)

type attrsKey struct{}

// ctxWithAttrs appends request-scoped attrs to the context. They are injected
// into every record emitted with this context by the contextHandler.
func ctxWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing := attrsFromCtx(ctx)

	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)

	return context.WithValue(ctx, attrsKey{}, merged)
}

func attrsFromCtx(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(attrsKey{}).([]slog.Attr)

	return attrs
}

// FromCtx returns the global logger. Request-scoped attrs stored in ctx are
// injected by the contextHandler at emit time, so the returned logger need not
// carry them itself.
func FromCtx(_ context.Context) *slog.Logger { return Log() }

// NewCtx stores attrs in ctx and returns it with a logger (the global logger).
func NewCtx(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	ctx = ctxWithAttrs(ctx, attrs...)

	return ctx, Log()
}

// CtxWithFields appends attrs to ctx and returns a logger that will emit them.
func CtxWithFields(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	return NewCtx(ctx, attrs...)
}

// WrapCtx appends attrs produced by wrap to ctx. Kept for call-site
// compatibility with the previous API shape.
func WrapCtx(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	return NewCtx(ctx, attrs...)
}

// WithIndent calls fn with a logger that prepends indent to every message.
// Nesting accumulates indents.
func WithIndent(l *slog.Logger, indent string, fn func(*slog.Logger)) {
	fn(slog.New(&indentHandler{indent: indent, next: l.Handler()}))
}
