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
	if ctx == nil {
		return nil
	}

	attrs, _ := ctx.Value(attrsKey{}).([]slog.Attr)

	return attrs
}

// WithContext binds ctx's request-scoped attrs to l so they are emitted on
// every record, even when the caller uses the non-Context methods
// (logger.Debug/Info/...) which slog dispatches with context.Background().
//
// slog.LogValuer attrs stay unresolved until a record is actually emitted, so
// this is safe to call on the hot path. When ctx carries no attrs, l is
// returned unchanged (no allocation).
func WithContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	if len(attrsFromCtx(ctx)) == 0 {
		return l
	}

	next := l.Handler()
	if ch, ok := next.(*contextHandler); ok {
		// Strip the existing (unbound) contextHandler so its per-call injection
		// does not run in addition to the bound one.
		next = ch.next
	}

	return slog.New(&contextHandler{next: next, bound: ctx})
}

// FromCtx returns the global logger bound to ctx's request-scoped attrs.
func FromCtx(ctx context.Context) *slog.Logger { return WithContext(ctx, Log()) }

// CtxWithFields appends attrs to ctx and returns the ctx together with a logger
// that emits them. The attrs are injected by the contextHandler at emit time.
func CtxWithFields(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	ctx = ctxWithAttrs(ctx, attrs...)

	return ctx, FromCtx(ctx)
}

// WithIndent calls fn with a logger that prepends indent to every message.
// Nesting accumulates indents.
func WithIndent(l *slog.Logger, indent string, fn func(*slog.Logger)) {
	fn(slog.New(&indentHandler{indent: indent, next: l.Handler()}))
}
