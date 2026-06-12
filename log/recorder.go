package log

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

// Recorder is a slog.Handler that records emitted records for test assertions.
// It replaces the old logrus MockLoggerHook / test.NewGlobal patterns.
type Recorder struct {
	mu      sync.Mutex
	records []slog.Record
}

// NewRecorder returns a logger writing into a fresh Recorder (trace level).
func NewRecorder() (*slog.Logger, *Recorder) {
	rec := &Recorder{}

	return slog.New(rec), rec
}

func (r *Recorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *Recorder) Handle(_ context.Context, rec slog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec.Clone())

	return nil
}

func (r *Recorder) WithAttrs([]slog.Attr) slog.Handler { return r }
func (r *Recorder) WithGroup(string) slog.Handler      { return r }

// Records returns a copy of the recorded records.
func (r *Recorder) Records() []slog.Record {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]slog.Record(nil), r.records...)
}

// Messages returns the recorded messages in order.
func (r *Recorder) Messages() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	msgs := make([]string, len(r.records))
	for i, rec := range r.records {
		msgs[i] = rec.Message
	}

	return msgs
}

// LastMessage returns the most recent message, or "" if none.
func (r *Recorder) LastMessage() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.records) == 0 {
		return ""
	}

	return r.records[len(r.records)-1].Message
}

// Attr looks up an attr by key on the most recent record.
func (r *Recorder) Attr(key string) (slog.Value, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.records) == 0 {
		return slog.Value{}, false
	}

	var found slog.Value

	var ok bool

	r.records[len(r.records)-1].Attrs(func(a slog.Attr) bool {
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = nil
}

// CaptureGlobal swaps the global logger for a Recorder and returns a restore
// func (call via DeferCleanup). Ginkgo runs specs serially per process, so this
// is safe within a spec.
func CaptureGlobal() (*Recorder, func()) {
	prev := logger
	rec := &Recorder{}
	logger = slog.New(rec)
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
