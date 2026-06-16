package querylog

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/0xERR0R/blocky/log"
)

const loggerPrefixLoggerWriter = "queryLog"

type LoggerWriter struct {
	logger *slog.Logger
}

func NewLoggerWriter() *LoggerWriter {
	return &LoggerWriter{logger: log.PrefixedLog(loggerPrefixLoggerWriter)}
}

func (d *LoggerWriter) Write(entry *LogEntry) {
	d.logger.LogAttrs(context.Background(), slog.LevelInfo, "query resolved", LogEntryFields(entry)...)
}

func (d *LoggerWriter) CleanUp() {
	// Nothing to do
}

// LogEntryFields returns the entry as flat, snake_case slog attrs, omitting
// zero-valued fields (matching the historical logrus WithFields output).
func LogEntryFields(entry *LogEntry) []slog.Attr {
	return withoutZeroes(
		slog.String("client_ip", entry.ClientIP),
		slog.String("client_names", strings.Join(entry.ClientNames, "; ")),
		slog.String("response_reason", entry.ResponseReason),
		slog.String("response_type", entry.ResponseType),
		slog.String("response_code", entry.ResponseCode),
		slog.String("question_name", entry.QuestionName),
		slog.String("question_type", entry.QuestionType),
		slog.String("answer", entry.Answer),
		slog.Int64("duration_ms", entry.DurationMs),
		slog.String("instance", entry.BlockyInstance),
	)
}

func withoutZeroes(attrs ...slog.Attr) []slog.Attr {
	result := attrs[:0]

	for _, a := range attrs {
		if !reflect.ValueOf(a.Value.Any()).IsZero() {
			result = append(result, a)
		}
	}

	return result
}
