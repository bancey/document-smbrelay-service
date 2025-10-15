package logger

import (
	"log"
	"os"
	"strings"
)

var (
	// Log levels
	DEBUG = 0
	INFO  = 1
	WARN  = 2
	ERROR = 3

	currentLevel = INFO
)

func init() {
	// Configure logging based on LOG_LEVEL environment variable
	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
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

	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Logging configured with level: %s", logLevel)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if currentLevel <= DEBUG {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if currentLevel <= INFO {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if currentLevel <= WARN {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if currentLevel <= ERROR {
		log.Printf("[ERROR] "+format, v...)
	}
}
