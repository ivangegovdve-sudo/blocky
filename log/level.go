package log

import (
	"fmt"
	"log/slog"
	"strings"
)

// LevelTrace is a custom slog level below Debug for very verbose, per-request
// tracing. slog supports custom levels as integer offsets (see slog docs).
const LevelTrace = slog.LevelDebug - 4

// Level is blocky's log level. It wraps slog.Level but parses the historical
// (logrus) level strings from YAML so existing configs keep working.
type Level slog.Level

// ToSlogLevel returns the underlying slog.Level.
func (l Level) ToSlogLevel() slog.Level { return slog.Level(l) }

var levelByName = map[string]slog.Level{
	"trace":   LevelTrace,
	"debug":   slog.LevelDebug,
	"info":    slog.LevelInfo,
	"warn":    slog.LevelWarn,
	"warning": slog.LevelWarn,
	"error":   slog.LevelError,
	"fatal":   slog.LevelError, // legacy: no code emits at fatal anymore
	"panic":   slog.LevelError, // legacy
}

var nameByLevel = map[slog.Level]string{
	LevelTrace:      "trace",
	slog.LevelDebug: "debug",
	slog.LevelInfo:  "info",
	slog.LevelWarn:  "warn",
	slog.LevelError: "error",
}

// String returns the canonical lower-case level name.
func (l Level) String() string {
	if name, ok := nameByLevel[slog.Level(l)]; ok {
		return name
	}

	return slog.Level(l).String()
}

// UnmarshalText implements encoding.TextUnmarshaler (used by the YAML loader,
// the same mechanism FormatType relies on).
func (l *Level) UnmarshalText(text []byte) error {
	name := strings.ToLower(strings.TrimSpace(string(text)))

	lvl, ok := levelByName[name]
	if !ok {
		return fmt.Errorf("invalid log level %q", string(text))
	}

	*l = Level(lvl)

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}
