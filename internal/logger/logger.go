package logger

import (
	"io"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	setupLogger()
}

func setupLogger() {
	level := getLogLevel()
	output := getLogOutput()

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: level,
	})

	defaultLogger = slog.New(handler)
}

func getLogLevel() slog.Level {
	env := os.Getenv("KEYP_LOG_LEVEL")

	levelMap := map[string]slog.Level{
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}

	if level, exists := levelMap[env]; exists {
		return level
	}

	if isTestEnvironment() {
		return slog.LevelError
	}

	return slog.LevelInfo
}

func getLogOutput() io.Writer {
	if isTestEnvironment() {
		return io.Discard
	}

	return os.Stdout
}

func isTestEnvironment() bool {
	return os.Getenv("KEYP_TEST_MODE") == "true"
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}
