package log

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

// recorderStore is the shared backing storage for a Recorder and any
// WithAttrs-derived recorders, so the original handle sees all records.
type recorderStore struct {
	mu      sync.Mutex
	records []slog.Record
}

// Recorder is a slog.Handler that records emitted records for test assertions.
// It replaces the old logrus MockLoggerHook / test.NewGlobal patterns.
type Recorder struct {
	store *recorderStore
	base  []slog.Attr // attrs pre-attached via WithAttrs
}

// NewRecorder returns a logger writing into a fresh Recorder (trace level).
func NewRecorder() (*slog.Logger, *Recorder) {
	rec := &Recorder{store: &recorderStore{}}

	return slog.New(rec), rec
}

func (r *Recorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *Recorder) Handle(_ context.Context, rec slog.Record) error {
	c := rec.Clone()
	if len(r.base) > 0 {
		// prepend base attrs (pre-attached via WithAttrs)
		c.AddAttrs(r.base...)
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.records = append(r.store.records, c)

	return nil
}

func (r *Recorder) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(r.base)+len(attrs))
	merged = append(merged, r.base...)
	merged = append(merged, attrs...)

	return &Recorder{store: r.store, base: merged}
}

func (r *Recorder) WithGroup(string) slog.Handler { return r }

// Records returns a copy of the recorded records.
func (r *Recorder) Records() []slog.Record {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	return append([]slog.Record(nil), r.store.records...)
}

// Messages returns the recorded messages in order.
func (r *Recorder) Messages() []string {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	msgs := make([]string, len(r.store.records))
	for i, rec := range r.store.records {
		msgs[i] = rec.Message
	}

	return msgs
}

// LastMessage returns the most recent message, or "" if none.
func (r *Recorder) LastMessage() string {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if len(r.store.records) == 0 {
		return ""
	}

	return r.store.records[len(r.store.records)-1].Message
}

// Attr looks up an attr by key on the most recent record.
func (r *Recorder) Attr(key string) (slog.Value, bool) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if len(r.store.records) == 0 {
		return slog.Value{}, false
	}

	var found slog.Value

	var ok bool

	r.store.records[len(r.store.records)-1].Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			found, ok = a.Value, true

			return false
		}

		return true
	})

	return found, ok
}

// Reset clears recorded records.
func (r *Recorder) Reset() {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.records = nil
}

// CaptureGlobal swaps the global logger for a Recorder (wrapped in a
// contextHandler so context-stored attrs are captured) and returns a restore
// func (call via DeferCleanup). Ginkgo runs specs serially per process, so this
// is safe within a spec.
func CaptureGlobal() (*Recorder, func()) {
	prev := logger
	rec := &Recorder{store: &recorderStore{}}
	logger = slog.New(&contextHandler{next: rec})
	slog.SetDefault(logger)

	return rec, func() {
		logger = prev
		slog.SetDefault(prev)
	}
}

// ConfigureForTest routes the global logger to w (typically GinkgoWriter):
// quiet on passing specs, full diagnostics on failure. No color, trace level.
func ConfigureForTest(w io.Writer) {
	levelVar.Set(LevelTrace)
	logger = slog.New(&contextHandler{next: slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: levelVar,
	})})
	slog.SetDefault(logger)
}
