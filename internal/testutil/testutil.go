package testutil

import (
	"os"
	"testing"
)

// SetupTestEnv sets up a clean test environment with SMB configuration
func SetupTestEnv(t *testing.T) func() {
	// Save original environment
	originalEnv := os.Environ()

	// Clear environment
	os.Clearenv()

	// Set test SMB configuration
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")
	os.Setenv("SMB_PORT", "445")
	os.Setenv("LOG_LEVEL", "ERROR") // Reduce noise during tests

	// Return cleanup function
	return func() {
		os.Clearenv()
		for _, env := range originalEnv {
			// Parse KEY=VALUE
			for i := 0; i < len(env); i++ {
				if env[i] == '=' {
					os.Setenv(env[:i], env[i+1:])
					break
				}
			}
		}
	}
}

// SetupEmptyEnv clears all environment variables for testing missing config
func SetupEmptyEnv(t *testing.T) func() {
	originalEnv := os.Environ()

	os.Clearenv()

	return func() {
		os.Clearenv()
		for _, env := range originalEnv {
			for i := 0; i < len(env); i++ {
				if env[i] == '=' {
					os.Setenv(env[:i], env[i+1:])
					break
				}
			}
		}
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, str, substr string) {
	t.Helper()
	if !contains(str, substr) {
		t.Errorf("Expected string to contain '%s', got: %s", substr, str)
	}
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string) {
	t.Helper()
	if contains(str, substr) {
		t.Errorf("Expected string to not contain '%s', got: %s", substr, str)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Errorf("Expected values to be different, both are %v", expected)
	}
}

// AssertTrue checks if a boolean is true
func AssertTrue(t *testing.T, value bool) {
	t.Helper()
	if !value {
		t.Error("Expected value to be true")
	}
}

// AssertFalse checks if a boolean is false
func AssertFalse(t *testing.T, value bool) {
	t.Helper()
	if value {
		t.Error("Expected value to be false")
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Error("Expected non-nil value")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
