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

func TestLoggerProductionMode(t *testing.T) {
	// Temporarily unset test mode
	originalTestMode := os.Getenv("KEYP_TEST_MODE")
	os.Unsetenv("KEYP_TEST_MODE")
	defer func() {
		if originalTestMode != "" {
			os.Setenv("KEYP_TEST_MODE", originalTestMode)
		}
	}()

	// Set a specific log level
	os.Setenv("KEYP_LOG_LEVEL", "ERROR")

	// These should work in production mode (but we can't easily test output)
	logger.Error("test error in production mode")

	// Restore test mode
	os.Setenv("KEYP_TEST_MODE", "true")
}

func TestLoggerInvalidLevel(t *testing.T) {
	// Set invalid log level
	os.Setenv("KEYP_LOG_LEVEL", "INVALID")
	os.Setenv("KEYP_TEST_MODE", "true")

	// Should default to INFO level and not panic
	logger.Info("test with invalid level")
}
