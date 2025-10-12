package log

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Init setup logger stuff.
// level can be error, warn, info, debug if something other supplied info level selected.
// fileDescriptor should be opened before supplying it to Init().
func Init(level string, fileDescriptor *os.File) {
	var loglevel slog.Level

	// no panic, no trace.
	switch level {
	case "error":
		loglevel = slog.LevelError

	case "warn":
		loglevel = slog.LevelWarn

	case "info":
		loglevel = slog.LevelInfo

	case "debug":
		loglevel = slog.LevelDebug

	default:
		loglevel = slog.LevelInfo

	}

	opts := &slog.HandlerOptions{
		// Use the ReplaceAttr function on the handler options
		// to be able to replace any single attribute in the log output
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// check that we are handling the time key
			if a.Key != slog.TimeKey {
				return a
			}

			t := a.Value.Time()

			// change the value from a time.Time to a String
			// where the string has the correct time format.
			a.Value = slog.StringValue(t.Format(time.DateTime))

			return a
		},

		Level: loglevel,
	}

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				fileDescriptor,
				opts,
			),
		),
	)
}

// Error passes message directly to slog.Error().
func Error(message string) {
	slog.Error(message)
}

// Errorf logs message on error log level and allows to supply arguments in printf() style.
func Errorf(format string, a ...any) {
	slog.Error(fmt.Sprintf(format, a...))
}

// Warn passes message directly to slog.Warn().
func Warn(message string) {
	slog.Warn(message)
}

// Warnf logs message on warn log level and allows to supply arguments in printf() style.
func Warnf(format string, a ...any) {
	slog.Warn(fmt.Sprintf(format, a...))
}

// Info passes message directly to slog.Info().
func Info(message string) {
	slog.Info(message)
}

// Infof logs message on info log level and allows to supply arguments in printf() style.
func Infof(format string, a ...any) {
	slog.Info(fmt.Sprintf(format, a...))
}

// Debug passes message directly to slog.Debug().
func Debug(message string) {
	slog.Debug(message)
}

// Debugf logs message on debug log level and allows to supply arguments in printf() style.
func Debugf(format string, a ...any) {
	slog.Debug(fmt.Sprintf(format, a...))
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
