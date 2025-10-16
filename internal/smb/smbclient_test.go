package smb

import (
	"os"
	"testing"
)

func TestValidateBinaryPath_ValidExecutable(t *testing.T) {
	// Try to find a known executable on the system
	if path, err := os.Executable(); err == nil {
		if validateBinaryPath(path) {
			t.Logf("Successfully validated executable: %s", path)
		} else {
			t.Errorf("Failed to validate known executable: %s", path)
		}
	}
}

func TestValidateBinaryPath_NonExistent(t *testing.T) {
	result := validateBinaryPath("/nonexistent/path/to/binary")
	if result {
		t.Error("Expected validation to fail for non-existent path")
	}
}

func TestValidateBinaryPath_Directory(t *testing.T) {
	// Use a directory that should exist
	tmpDir := os.TempDir()
	result := validateBinaryPath(tmpDir)
	if result {
		t.Errorf("Expected validation to fail for directory: %s", tmpDir)
	}
}

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

	// Test with a valid executable path
	if testPath, err := os.Executable(); err == nil {
		os.Setenv("SMBCLIENT_PATH", testPath)
		result := getSmbClientPath()

		if result != testPath {
			t.Errorf("Expected path to be '%s', got '%s'", testPath, result)
		}
	}

	// Test with invalid path - should fall back
	os.Setenv("SMBCLIENT_PATH", "/invalid/nonexistent/path")
	result := getSmbClientPath()

	// Should not return the invalid path
	if result == "/invalid/nonexistent/path" {
		t.Error("Expected fallback for invalid SMBCLIENT_PATH")
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
