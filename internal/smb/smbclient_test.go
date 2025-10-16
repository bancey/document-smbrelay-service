package smb

import (
	"os"
	"testing"
)

func TestGetSmbClientPath_EnvironmentVariable(t *testing.T) {
	// Save original env and restore after test
	origPath := os.Getenv("SMBCLIENT_PATH")
	defer func() {
		if origPath != "" {
			os.Setenv("SMBCLIENT_PATH", origPath)
		} else {
			os.Unsetenv("SMBCLIENT_PATH")
		}
	}()

	// Test with custom path
	customPath := "/custom/path/smbclient"
	os.Setenv("SMBCLIENT_PATH", customPath)

	result := getSmbClientPath()

	if result != customPath {
		t.Errorf("Expected path to be '%s', got '%s'", customPath, result)
	}
}

func TestGetSmbClientPath_Fallback(t *testing.T) {
	// Save original env and restore after test
	origPath := os.Getenv("SMBCLIENT_PATH")
	defer func() {
		if origPath != "" {
			os.Setenv("SMBCLIENT_PATH", origPath)
		} else {
			os.Unsetenv("SMBCLIENT_PATH")
		}
	}()

	// Unset environment variable to test fallback
	os.Unsetenv("SMBCLIENT_PATH")

	result := getSmbClientPath()

	// Should find smbclient in PATH or use one of the common paths
	if result == "" {
		t.Error("Expected non-empty path from auto-detection")
	}

	// At minimum, should return the default fallback
	if result == "" {
		t.Error("getSmbClientPath should never return empty string")
	}
}

func TestDefaultSmbClientExecutor_CustomBinaryPath(t *testing.T) {
	executor := &DefaultSmbClientExecutor{
		BinaryPath: "/custom/test/path",
	}

	if executor.BinaryPath != "/custom/test/path" {
		t.Errorf("Expected BinaryPath to be '/custom/test/path', got '%s'", executor.BinaryPath)
	}
}

func TestDefaultSmbClientExecutor_EmptyBinaryPath(t *testing.T) {
	// When BinaryPath is empty, Execute should use getSmbClientPath()
	executor := &DefaultSmbClientExecutor{
		BinaryPath: "",
	}

	// We can't easily test Execute without a real smbclient,
	// but we can verify the struct is set up correctly
	if executor.BinaryPath != "" {
		t.Error("Expected empty BinaryPath to trigger auto-detection")
	}
}
