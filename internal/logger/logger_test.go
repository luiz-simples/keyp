package logger_test

import (
	"os"
	"testing"

	"github.com/luiz-simples/keyp.git/internal/logger"
)

func TestLoggerInTestMode(t *testing.T) {
	os.Setenv("KEYP_TEST_MODE", "true")

	logger.Info("test info message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")
	logger.Error("test error message")

}

func TestLoggerLevels(t *testing.T) {
	testCases := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range testCases {
		os.Setenv("KEYP_LOG_LEVEL", level)
		os.Setenv("KEYP_TEST_MODE", "true") // Keep test mode to avoid output

		logger.Info("test message for level", "level", level)
	}
}

func TestLoggerProductionMode(t *testing.T) {
	originalTestMode := os.Getenv("KEYP_TEST_MODE")
	os.Unsetenv("KEYP_TEST_MODE")
	defer func() {
		if originalTestMode != "" {
			os.Setenv("KEYP_TEST_MODE", originalTestMode)
		}
	}()

	os.Setenv("KEYP_LOG_LEVEL", "ERROR")

	logger.Error("test error in production mode")

	os.Setenv("KEYP_TEST_MODE", "true")
}

func TestLoggerInvalidLevel(t *testing.T) {
	os.Setenv("KEYP_LOG_LEVEL", "INVALID")
	os.Setenv("KEYP_TEST_MODE", "true")

	logger.Info("test with invalid level")
}
