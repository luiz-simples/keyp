package logger

import (
	"os"
)

func HasError(err error) bool {
	return err != nil
}

func IsEmpty(data string) bool {
	return len(data) == 0
}

func isValidLevel(level string) bool {
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}

	return false
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if IsEmpty(value) {
		return defaultValue
	}
	return value
}
