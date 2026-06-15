package querylog

import (
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
	d.logger.Info("query resolved", slog.Any("entry", entry))
}

func (d *LoggerWriter) CleanUp() {
	// Nothing to do
}

func LogEntryFields(entry *LogEntry) map[string]any {
	return withoutZeroes(map[string]any{
		"client_ip":       entry.ClientIP,
		"client_names":    strings.Join(entry.ClientNames, "; "),
		"response_reason": entry.ResponseReason,
		"response_type":   entry.ResponseType,
		"response_code":   entry.ResponseCode,
		"question_name":   entry.QuestionName,
		"question_type":   entry.QuestionType,
		"answer":          entry.Answer,
		"duration_ms":     entry.DurationMs,
		"instance":        entry.BlockyInstance,
	})
}

func withoutZeroes(fields map[string]any) map[string]any {
	for k, v := range fields {
		if reflect.ValueOf(v).IsZero() {
			delete(fields, k)
		}
	}

	return fields
}
