package logger

import (
	"os"
	"testing"
)

func TestHasError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"non-nil error", os.ErrNotExist, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasError(tt.err)
			if result != tt.expected {
				t.Errorf("hasError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"whitespace string", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.data)
			if result != tt.expected {
				t.Errorf("isEmpty(%q) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestIsValidLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"DEBUG level", "DEBUG", true},
		{"INFO level", "INFO", true},
		{"WARN level", "WARN", true},
		{"ERROR level", "ERROR", true},
		{"invalid level", "INVALID", false},
		{"lowercase level", "debug", false},
		{"empty level", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLevel(tt.level)
			if result != tt.expected {
				t.Errorf("isValidLevel(%q) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{"env var exists", "TEST_KEY", "default", "env_value", "env_value"},
		{"env var empty", "TEST_KEY", "default", "", "default"},
		{"env var not set", "NONEXISTENT_KEY", "default", "", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.key)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithDefault(%q, %q) = %q, want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		testMode string
		expected string
	}{
		{"DEBUG level", "DEBUG", "", "DEBUG"},
		{"INFO level", "INFO", "", "INFO"},
		{"WARN level", "WARN", "", "WARN"},
		{"ERROR level", "ERROR", "", "ERROR"},
		{"invalid level in test mode", "INVALID", "true", "ERROR"},
		{"invalid level in production", "INVALID", "", "INFO"},
		{"empty level in test mode", "", "true", "ERROR"},
		{"empty level in production", "", "", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv("KEYP_LOG_LEVEL")
			os.Unsetenv("KEYP_TEST_MODE")

			// Set environment variables
			if tt.envValue != "" {
				os.Setenv("KEYP_LOG_LEVEL", tt.envValue)
				defer os.Unsetenv("KEYP_LOG_LEVEL")
			}
			if tt.testMode != "" {
				os.Setenv("KEYP_TEST_MODE", tt.testMode)
				defer os.Unsetenv("KEYP_TEST_MODE")
			}

			result := getLogLevel()
			if result.String() != tt.expected {
				t.Errorf("getLogLevel() = %v, want %v", result.String(), tt.expected)
			}
		})
	}
}

func TestGetLogOutput(t *testing.T) {
	tests := []struct {
		name      string
		testMode  string
		isDiscard bool
	}{
		{"test mode enabled", "true", true},
		{"test mode disabled", "", false},
		{"test mode false", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv("KEYP_TEST_MODE")

			// Set environment variable
			if tt.testMode != "" {
				os.Setenv("KEYP_TEST_MODE", tt.testMode)
				defer os.Unsetenv("KEYP_TEST_MODE")
			}

			result := getLogOutput()

			// Check if it's io.Discard or os.Stdout
			if tt.isDiscard {
				// We can't directly compare io.Discard, but we can test by writing to it
				n, err := result.Write([]byte("test"))
				if err != nil || n != 4 {
					t.Errorf("getLogOutput() in test mode should return io.Discard-like writer")
				}
			} else {
				if result != os.Stdout {
					t.Errorf("getLogOutput() in production mode should return os.Stdout")
				}
			}
		})
	}
}
