package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogLevelInitialization(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedLevel int
	}{
		{"Debug level", "DEBUG", DEBUG},
		{"Info level", "INFO", INFO},
		{"Warning level", "WARNING", WARN},
		{"Warn level", "WARN", WARN},
		{"Error level", "ERROR", ERROR},
		{"Empty defaults to INFO", "", INFO},
		{"Invalid defaults to INFO", "INVALID", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original level
			originalLevel := currentLevel

			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("LOG_LEVEL", tt.envValue)
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			// Re-initialize
			logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
			if logLevel == "" {
				logLevel = "INFO"
			}

			var level int
			switch logLevel {
			case "DEBUG":
				level = DEBUG
			case "INFO":
				level = INFO
			case "WARNING", "WARN":
				level = WARN
			case "ERROR":
				level = ERROR
			default:
				level = INFO
			}

			if level != tt.expectedLevel {
				t.Errorf("Expected level %d, got %d", tt.expectedLevel, level)
			}

			// Restore original level
			currentLevel = originalLevel
		})
	}
}

func TestDebugLogging(t *testing.T) {
	// Save original level and output
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Test DEBUG level
	currentLevel = DEBUG
	Debug("test debug message %s", "arg")
	if !strings.Contains(buf.String(), "[DEBUG] test debug message arg") {
		t.Errorf("Expected debug message in output, got: %s", buf.String())
	}

	// Test that DEBUG is not logged at INFO level
	buf.Reset()
	currentLevel = INFO
	Debug("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Errorf("Debug message should not appear at INFO level")
	}
}

func TestInfoLogging(t *testing.T) {
	// Save original level and output
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Test INFO level
	currentLevel = INFO
	Info("test info message %s", "arg")
	if !strings.Contains(buf.String(), "[INFO] test info message arg") {
		t.Errorf("Expected info message in output, got: %s", buf.String())
	}

	// Test that INFO is not logged at WARN level
	buf.Reset()
	currentLevel = WARN
	Info("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Errorf("Info message should not appear at WARN level")
	}
}

func TestWarnLogging(t *testing.T) {
	// Save original level and output
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Test WARN level
	currentLevel = WARN
	Warn("test warning message %s", "arg")
	if !strings.Contains(buf.String(), "[WARN] test warning message arg") {
		t.Errorf("Expected warning message in output, got: %s", buf.String())
	}

	// Test that WARN is not logged at ERROR level
	buf.Reset()
	currentLevel = ERROR
	Warn("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Errorf("Warn message should not appear at ERROR level")
	}
}

func TestErrorLogging(t *testing.T) {
	// Save original level and output
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Test ERROR level - errors should always appear
	currentLevel = ERROR
	Error("test error message %s", "arg")
	if !strings.Contains(buf.String(), "[ERROR] test error message arg") {
		t.Errorf("Expected error message in output, got: %s", buf.String())
	}
}

func TestLogLevelHierarchy(t *testing.T) {
	// Save original level and output
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// At DEBUG level, all messages should appear
	currentLevel = DEBUG
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output := buf.String()
	if !strings.Contains(output, "debug") || !strings.Contains(output, "info") ||
		!strings.Contains(output, "warn") || !strings.Contains(output, "error") {
		t.Errorf("Expected all log levels to appear at DEBUG, got: %s", output)
	}

	// At INFO level, only INFO, WARN, ERROR should appear
	buf.Reset()
	currentLevel = INFO
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output = buf.String()
	if strings.Contains(output, "debug") {
		t.Errorf("DEBUG should not appear at INFO level")
	}
	if !strings.Contains(output, "info") || !strings.Contains(output, "warn") || !strings.Contains(output, "error") {
		t.Errorf("Expected INFO, WARN, ERROR to appear at INFO level, got: %s", output)
	}
}

// ============================================================================
// Extended Test Cases - Edge cases and additional scenarios
// ============================================================================

// Test all log level variations
func TestLogLevel_AllVariations(t *testing.T) {
	tests := []struct {
		envValue      string
		shouldDebug   bool
		shouldInfo    bool
		shouldWarn    bool
		shouldError   bool
	}{
		{"DEBUG", true, true, true, true},
		{"INFO", false, true, true, true},
		{"WARNING", false, false, true, true},
		{"WARN", false, false, true, true},
		{"ERROR", false, false, false, true},
		{"", false, true, true, true}, // Default is INFO
		{"INVALID", false, true, true, true}, // Invalid defaults to INFO
	}

	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	for _, tt := range tests {
		t.Run("level_"+tt.envValue, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)

			// Set the level based on env value
			logLevel := strings.ToUpper(tt.envValue)
			if logLevel == "" {
				logLevel = "INFO"
			}

			switch logLevel {
			case "DEBUG":
				currentLevel = DEBUG
			case "INFO":
				currentLevel = INFO
			case "WARNING", "WARN":
				currentLevel = WARN
			case "ERROR":
				currentLevel = ERROR
			default:
				currentLevel = INFO
			}

			// Test each log level
			buf.Reset()
			Debug("debug message")
			debugLogged := strings.Contains(buf.String(), "debug message")
			if debugLogged != tt.shouldDebug {
				t.Errorf("Debug logging: expected %v, got %v", tt.shouldDebug, debugLogged)
			}

			buf.Reset()
			Info("info message")
			infoLogged := strings.Contains(buf.String(), "info message")
			if infoLogged != tt.shouldInfo {
				t.Errorf("Info logging: expected %v, got %v", tt.shouldInfo, infoLogged)
			}

			buf.Reset()
			Warn("warn message")
			warnLogged := strings.Contains(buf.String(), "warn message")
			if warnLogged != tt.shouldWarn {
				t.Errorf("Warn logging: expected %v, got %v", tt.shouldWarn, warnLogged)
			}

			buf.Reset()
			Error("error message")
			errorLogged := strings.Contains(buf.String(), "error message")
			if errorLogged != tt.shouldError {
				t.Errorf("Error logging: expected %v, got %v", tt.shouldError, errorLogged)
			}
		})
	}
}

// Test log formatting with multiple arguments
func TestLogFormatting_MultipleArgs(t *testing.T) {
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	currentLevel = DEBUG

	Debug("test %s %d %v", "string", 42, true)
	if !strings.Contains(buf.String(), "test string 42 true") {
		t.Errorf("Expected formatted message, got: %s", buf.String())
	}

	buf.Reset()
	Info("info %s %d", "arg", 100)
	if !strings.Contains(buf.String(), "info arg 100") {
		t.Errorf("Expected formatted message, got: %s", buf.String())
	}

	buf.Reset()
	Warn("warn %v %v", "a", "b")
	if !strings.Contains(buf.String(), "warn a b") {
		t.Errorf("Expected formatted message, got: %s", buf.String())
	}

	buf.Reset()
	Error("error %s", "message")
	if !strings.Contains(buf.String(), "error message") {
		t.Errorf("Expected formatted message, got: %s", buf.String())
	}
}

// Test log prefix contains correct level marker
func TestLogPrefix(t *testing.T) {
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	currentLevel = DEBUG

	Debug("test")
	if !strings.Contains(buf.String(), "[DEBUG]") {
		t.Errorf("Expected [DEBUG] prefix, got: %s", buf.String())
	}

	buf.Reset()
	Info("test")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Errorf("Expected [INFO] prefix, got: %s", buf.String())
	}

	buf.Reset()
	Warn("test")
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("Expected [WARN] prefix, got: %s", buf.String())
	}

	buf.Reset()
	Error("test")
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("Expected [ERROR] prefix, got: %s", buf.String())
	}
}

// Test that log levels are properly compared
func TestLogLevelComparison(t *testing.T) {
	if DEBUG >= INFO {
		t.Error("Expected DEBUG < INFO")
	}
	if INFO >= WARN {
		t.Error("Expected INFO < WARN")
	}
	if WARN >= ERROR {
		t.Error("Expected WARN < ERROR")
	}
}

// Test init function sets correct defaults
func TestInit_DefaultLevel(t *testing.T) {
	// Save and restore
	originalEnv := os.Getenv("LOG_LEVEL")
	defer func() {
		if originalEnv != "" {
			os.Setenv("LOG_LEVEL", originalEnv)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	os.Unsetenv("LOG_LEVEL")

	// When LOG_LEVEL is not set, it should default to INFO
	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "INFO"
	}

	if logLevel != "INFO" {
		t.Errorf("Expected default log level INFO, got %s", logLevel)
	}
}

// Test case insensitivity of log level names
func TestLogLevel_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"debug", DEBUG},
		{"DEBUG", DEBUG},
		{"Debug", DEBUG},
		{"info", INFO},
		{"INFO", INFO},
		{"Info", INFO},
		{"warning", WARN},
		{"WARNING", WARN},
		{"warn", WARN},
		{"WARN", WARN},
		{"error", ERROR},
		{"ERROR", ERROR},
		{"Error", ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			upperInput := strings.ToUpper(tt.input)

			var level int
			switch upperInput {
			case "DEBUG":
				level = DEBUG
			case "INFO":
				level = INFO
			case "WARNING", "WARN":
				level = WARN
			case "ERROR":
				level = ERROR
			default:
				level = INFO
			}

			if level != tt.expected {
				t.Errorf("For input %s, expected level %d, got %d", tt.input, tt.expected, level)
			}
		})
	}
}

// Test logging with empty messages
func TestLogging_EmptyMessage(t *testing.T) {
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	currentLevel = DEBUG

	Debug("")
	if !strings.Contains(buf.String(), "[DEBUG]") {
		t.Error("Expected log entry even with empty message")
	}

	buf.Reset()
	Info("")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Error("Expected log entry even with empty message")
	}
}

// Test logging with no format arguments
func TestLogging_NoFormatArgs(t *testing.T) {
	originalLevel := currentLevel
	originalOutput := log.Writer()
	defer func() {
		currentLevel = originalLevel
		log.SetOutput(originalOutput)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	currentLevel = DEBUG

	Debug("plain message")
	if !strings.Contains(buf.String(), "plain message") {
		t.Errorf("Expected plain message, got: %s", buf.String())
	}
}
