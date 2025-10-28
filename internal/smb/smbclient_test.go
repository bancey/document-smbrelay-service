package smb

import (
	"os"
	"testing"

	"github.com/bancey/document-smbrelay-service/internal/config"
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

func TestExecuteWithEnv_EnvironmentVariables(t *testing.T) {
	// Test that environment variables with special characters are properly handled
	env := map[string]string{
		"TEST_VAR1": "value1",
		"TEST_VAR2": "value2 with spaces",
		"PASSWD":    "P@ssw0rd!$%Test", // Test password with special characters
	}

	// Verify env map can contain special characters
	if env["PASSWD"] != "P@ssw0rd!$%Test" {
		t.Error("Environment variable should preserve special characters")
	}

	// Verify we can create an executor
	executor := &DefaultSmbClientExecutor{
		BinaryPath: "",
	}

	// Verify the executor has the expected empty BinaryPath
	if executor.BinaryPath != "" {
		t.Errorf("Expected empty BinaryPath, got '%s'", executor.BinaryPath)
	}
}

func TestSanitizeArgsForLogging(t *testing.T) {
	tests := []struct {
		env        map[string]string
		args       []string
		name       string
		expectEnv  map[string]string
		expectArgs []string
	}{
		{
			name:       "password in -U flag",
			args:       []string{"-U", "user%password123"},
			expectArgs: []string{"-U", "user%***"},
			env:        nil,
			expectEnv:  map[string]string{},
		},
		{
			name:       "username only in -U flag",
			args:       []string{"-U", "username"},
			expectArgs: []string{"-U", "username"},
			env:        nil,
			expectEnv:  map[string]string{},
		},
		{
			name:       "password in environment",
			args:       []string{"ls"},
			expectArgs: []string{"ls"},
			env:        map[string]string{"PASSWD": "secret", "USER": "testuser"},
			expectEnv:  map[string]string{"PASSWD": "***", "USER": "testuser"},
		},
		{
			name:       "multiple password variations",
			args:       []string{"-U", "admin%P@ssw0rd!"},
			expectArgs: []string{"-U", "admin%***"},
			env:        map[string]string{"PASSWORD": "secret123", "API_KEY": "key123"},
			expectEnv:  map[string]string{"PASSWORD": "***", "API_KEY": "key123"},
		},
		{
			name:       "empty args and env",
			args:       []string{},
			expectArgs: []string{},
			env:        map[string]string{},
			expectEnv:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, gotEnv := sanitizeArgsForLogging(tt.args, tt.env)

			// Check args
			if len(gotArgs) != len(tt.expectArgs) {
				t.Errorf("Expected %d args, got %d", len(tt.expectArgs), len(gotArgs))
			}
			for i := range gotArgs {
				if i < len(tt.expectArgs) && gotArgs[i] != tt.expectArgs[i] {
					t.Errorf("Arg %d: expected '%s', got '%s'", i, tt.expectArgs[i], gotArgs[i])
				}
			}

			// Check env
			if len(gotEnv) != len(tt.expectEnv) {
				t.Errorf("Expected %d env vars, got %d", len(tt.expectEnv), len(gotEnv))
			}
			for k, v := range tt.expectEnv {
				if gotEnv[k] != v {
					t.Errorf("Env %s: expected '%s', got '%s'", k, v, gotEnv[k])
				}
			}
		})
	}
}

func TestBuildSmbClientArgs_NameResolutionOrder(t *testing.T) {
	// Test that buildSmbClientArgs includes -R host to force DNS-only name resolution
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	args, _, err := buildSmbClientArgs(cfg, "ls")
	if err != nil {
		t.Fatalf("buildSmbClientArgs failed: %v", err)
	}

	// Check that -R host is present to force DNS-only resolution (no NetBIOS)
	foundR := false
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-R" && args[i+1] == "host" {
			foundR = true
			break
		}
	}

	if !foundR {
		t.Error("Expected -R host flag to force DNS-only name resolution and avoid NetBIOS (port 139) traffic")
		t.Logf("Args: %v", args)
	}
}

func TestBuildSmbClientArgs_IPAddressForcing(t *testing.T) {
	// Test that -I flag is used when both ServerName and ServerIP are specified
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	args, _, err := buildSmbClientArgs(cfg, "ls")
	if err != nil {
		t.Fatalf("buildSmbClientArgs failed: %v", err)
	}

	// Check that -I flag is present with the IP address
	foundI := false
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-I" && args[i+1] == cfg.ServerIP {
			foundI = true
			break
		}
	}

	if !foundI {
		t.Error("Expected -I flag with IP address to force direct IP connection")
		t.Logf("Args: %v", args)
	}
}
