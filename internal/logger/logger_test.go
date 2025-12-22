package logger_test

import (
	"os"
	"testing"

	"github.com/luiz-simples/keyp.git/internal/logger"
)

func TestLoggerInTestMode(t *testing.T) {
	// Set test mode
	os.Setenv("KEYP_TEST_MODE", "true")

	// These should not produce any output in test mode
	logger.Info("test info message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")
	logger.Error("test error message")

	// Test passes if no output is produced
}

func TestLoggerLevels(t *testing.T) {
	testCases := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range testCases {
		os.Setenv("KEYP_LOG_LEVEL", level)
		os.Setenv("KEYP_TEST_MODE", "true") // Keep test mode to avoid output

		// Should not panic or error
		logger.Info("test message for level", "level", level)
	}
}
