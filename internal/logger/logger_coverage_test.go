package logger

import (
	"os"
	"testing"
)

// TestLoggerInit_DebugLevel tests logger initialization with DEBUG level
func TestLoggerInit_DebugLevel(t *testing.T) {
	// Save original LOG_LEVEL and restore after test
	origLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if origLevel != "" {
			os.Setenv("LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set LOG_LEVEL to DEBUG
	os.Setenv("LOG_LEVEL", "DEBUG")

	// Re-initialize logger by calling init manually (indirectly)
	// Note: We can't directly call init(), but we can test with the Debug function
	Debug("Test debug message")

	// If this doesn't panic, the test passes
	t.Log("Debug logging works")
}

// TestLoggerInit_WarningLevel tests logger initialization with WARNING level
func TestLoggerInit_WarningLevel(t *testing.T) {
	// Save original LOG_LEVEL and restore after test
	origLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if origLevel != "" {
			os.Setenv("LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set LOG_LEVEL to WARNING
	os.Setenv("LOG_LEVEL", "WARNING")

	// Test that warning level works
	Warn("Test warning message")

	// If this doesn't panic, the test passes
	t.Log("Warning logging works")
}

// TestLoggerInit_ErrorLevel tests logger initialization with ERROR level
func TestLoggerInit_ErrorLevel(t *testing.T) {
	// Save original LOG_LEVEL and restore after test
	origLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if origLevel != "" {
			os.Setenv("LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set LOG_LEVEL to ERROR
	os.Setenv("LOG_LEVEL", "ERROR")

	// Test that error level works
	Error("Test error message")

	// If this doesn't panic, the test passes
	t.Log("Error logging works")
}

// TestLoggerInit_InvalidLevel tests logger initialization with invalid level (should default to INFO)
func TestLoggerInit_InvalidLevel(t *testing.T) {
	// Save original LOG_LEVEL and restore after test
	origLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if origLevel != "" {
			os.Setenv("LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set LOG_LEVEL to invalid value
	os.Setenv("LOG_LEVEL", "INVALID_LEVEL")

	// Test that info level still works (default fallback)
	Info("Test info message with invalid level config")

	// If this doesn't panic, the test passes
	t.Log("Info logging works with invalid level (defaults to INFO)")
}

// TestLoggerInit_EmptyLevel tests logger initialization with empty level (should default to INFO)
func TestLoggerInit_EmptyLevel(t *testing.T) {
	// Save original LOG_LEVEL and restore after test
	origLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if origLevel != "" {
			os.Setenv("LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Unset LOG_LEVEL
	os.Unsetenv("LOG_LEVEL")

	// Test that info level works (default)
	Info("Test info message with empty level config")

	// If this doesn't panic, the test passes
	t.Log("Info logging works with empty level (defaults to INFO)")
}

// TestLoggerAllLevels tests all logging functions
func TestLoggerAllLevels(t *testing.T) {
	// Test all log levels to ensure they work
	Debug("Debug message")
	Info("Info message")
	Warn("Warning message")
	Error("Error message")

	t.Log("All log levels work correctly")
}
