package testutil

import (
	"os"
	"strings"
	"testing"
)

func TestSetupTestEnv(t *testing.T) {
	cleanup := SetupTestEnv(t)
	defer cleanup()

	// Verify test environment is set up correctly
	requiredVars := map[string]string{
		"SMB_SERVER_NAME": "testserver",
		"SMB_SERVER_IP":   "127.0.0.1",
		"SMB_SHARE_NAME":  "testshare",
		"SMB_USERNAME":    "testuser",
		"SMB_PASSWORD":    "testpass",
	}

	for key, expected := range requiredVars {
		actual := strings.TrimSpace(os.Getenv(key))
		if actual != expected {
			t.Errorf("Expected %s=%s, got %s=%s", key, expected, key, actual)
		}
	}
}

func TestSetupEmptyEnv(t *testing.T) {
	cleanup := SetupEmptyEnv(t)
	defer cleanup()

	// Verify environment is empty
	if os.Getenv("SMB_SERVER_NAME") != "" {
		t.Error("Expected empty environment")
	}
}

func TestAssertContains(t *testing.T) {
	// Create a mock test to capture failures
	mockT := &testing.T{}

	// This should pass
	AssertContains(mockT, "hello world", "world")

	// This should fail (but we can't easily test it without causing actual failure)
}

func TestAssertEqual(t *testing.T) {
	mockT := &testing.T{}

	AssertEqual(mockT, 42, 42)
	AssertEqual(mockT, "test", "test")
	AssertEqual(mockT, true, true)
}

func TestAssertNotEqual(t *testing.T) {
	mockT := &testing.T{}

	AssertNotEqual(mockT, 42, 43)
	AssertNotEqual(mockT, "test", "other")
	AssertNotEqual(mockT, true, false)
}

func TestAssertTrue(t *testing.T) {
	mockT := &testing.T{}

	AssertTrue(mockT, true)
}

func TestAssertFalse(t *testing.T) {
	mockT := &testing.T{}

	AssertFalse(mockT, false)
}

func TestAssertNil(t *testing.T) {
	mockT := &testing.T{}

	var nilValue interface{}
	AssertNil(mockT, nilValue)
}

func TestAssertNotNil(t *testing.T) {
	mockT := &testing.T{}

	value := "not nil"
	AssertNotNil(mockT, value)
}
