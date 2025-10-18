package smb

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

func TestIsRetryableError_ConnectionRefused(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		output   string
		expected bool
	}{
		{
			name:     "Connection refused in error",
			err:      errors.New("connection refused"),
			output:   "",
			expected: true,
		},
		{
			name:     "Connection refused in output",
			err:      errors.New("command failed"),
			output:   "Connection to 127.0.0.1 failed (Error NT_STATUS_CONNECTION_REFUSED)",
			expected: true,
		},
		{
			name:     "Timeout error",
			err:      errors.New("i/o timeout"),
			output:   "",
			expected: true,
		},
		{
			name:     "Network unreachable",
			err:      errors.New("network is unreachable"),
			output:   "",
			expected: true,
		},
		{
			name:     "NT_STATUS_IO_TIMEOUT",
			err:      errors.New("command failed"),
			output:   "NT_STATUS_IO_TIMEOUT",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err, tt.output)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError_NonRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		output   string
		expected bool
	}{
		{
			name:     "Authentication failure",
			err:      errors.New("command failed"),
			output:   "NT_STATUS_LOGON_FAILURE",
			expected: false,
		},
		{
			name:     "Access denied",
			err:      errors.New("command failed"),
			output:   "NT_STATUS_ACCESS_DENIED",
			expected: false,
		},
		{
			name:     "Bad network name",
			err:      errors.New("command failed"),
			output:   "NT_STATUS_BAD_NETWORK_NAME",
			expected: false,
		},
		{
			name:     "Object not found",
			err:      errors.New("command failed"),
			output:   "NT_STATUS_OBJECT_NAME_NOT_FOUND",
			expected: false,
		},
		{
			name:     "No error",
			err:      nil,
			output:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err, tt.output)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	cfg := &config.SMBConfig{
		InitialRetryDelay: 1.0,
		MaxRetryDelay:     30.0,
		RetryBackoff:      2.0,
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "First retry",
			attempt:  0,
			expected: 1 * time.Second,
		},
		{
			name:     "Second retry",
			attempt:  1,
			expected: 2 * time.Second,
		},
		{
			name:     "Third retry",
			attempt:  2,
			expected: 4 * time.Second,
		},
		{
			name:     "Fourth retry",
			attempt:  3,
			expected: 8 * time.Second,
		},
		{
			name:     "Fifth retry (capped)",
			attempt:  5,
			expected: 30 * time.Second, // Should be capped at MaxRetryDelay
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBackoff(tt.attempt, cfg)
			if result != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	cfg := &config.SMBConfig{
		MaxRetries:        3,
		InitialRetryDelay: 0.01, // Short delay for testing
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "success", nil
	}

	output, err := executeWithRetry("test operation", cfg, fn)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output != "success" {
		t.Errorf("Expected output 'success', got: %s", output)
	}
	if callCount != 1 {
		t.Errorf("Expected function to be called once, called %d times", callCount)
	}
}

func TestExecuteWithRetry_TransientErrorThenSuccess(t *testing.T) {
	cfg := &config.SMBConfig{
		MaxRetries:        3,
		InitialRetryDelay: 0.01, // Short delay for testing
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount < 3 {
			return "Connection refused", errors.New("connection refused")
		}
		return "success", nil
	}

	start := time.Now()
	output, err := executeWithRetry("test operation", cfg, fn)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output != "success" {
		t.Errorf("Expected output 'success', got: %s", output)
	}
	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, called %d times", callCount)
	}
	// Should have at least 2 delays (0.01s + 0.02s = 0.03s minimum)
	if elapsed < 30*time.Millisecond {
		t.Errorf("Expected at least 30ms elapsed for retries, got %v", elapsed)
	}
}

func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	cfg := &config.SMBConfig{
		MaxRetries:        3,
		InitialRetryDelay: 0.01, // Short delay for testing
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "NT_STATUS_ACCESS_DENIED", errors.New("access denied")
	}

	output, err := executeWithRetry("test operation", cfg, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != "NT_STATUS_ACCESS_DENIED" {
		t.Errorf("Expected output 'NT_STATUS_ACCESS_DENIED', got: %s", output)
	}
	if callCount != 1 {
		t.Errorf("Expected function to be called once (no retries for non-retryable errors), called %d times", callCount)
	}
}

func TestExecuteWithRetry_MaxRetriesExceeded(t *testing.T) {
	cfg := &config.SMBConfig{
		MaxRetries:        2, // Only 2 retries allowed
		InitialRetryDelay: 0.01,
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "Connection refused", errors.New("connection refused")
	}

	output, err := executeWithRetry("test operation", cfg, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != "Connection refused" {
		t.Errorf("Expected output 'Connection refused', got: %s", output)
	}
	// Should be called 3 times: initial + 2 retries
	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times (initial + 2 retries), called %d times", callCount)
	}
}

func TestExecuteWithRetry_ZeroRetries(t *testing.T) {
	cfg := &config.SMBConfig{
		MaxRetries:        0, // No retries
		InitialRetryDelay: 0.01,
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "Connection refused", errors.New("connection refused")
	}

	output, err := executeWithRetry("test operation", cfg, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != "Connection refused" {
		t.Errorf("Expected output 'Connection refused', got: %s", output)
	}
	// Should be called only once with no retries
	if callCount != 1 {
		t.Errorf("Expected function to be called once (no retries), called %d times", callCount)
	}
}

// Test integration with mock executor
func TestUploadFile_WithTransientError(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Create a mock that fails twice with transient error, then succeeds
	callCount := 0
	smbClientExec = &MockSmbClientExecutor{
		ExecuteFunc: func(args []string) (string, error) {
			callCount++
			if callCount < 3 {
				// Simulate transient connection error
				return "Connection to 127.0.0.1 failed", fmt.Errorf("connection refused")
			}
			// Check what command is being executed
			for i, arg := range args {
				if arg == "-c" && i+1 < len(args) {
					cmd := args[i+1]
					// Handle put command (upload)
					if contains(cmd, "put") {
						return "putting file test.txt as test.txt (1.0 kb/s) (average 1.0 kb/s)\n", nil
					}
					// Handle ls command (file existence check)
					if contains(cmd, "ls") {
						return "NT_STATUS_NO_SUCH_FILE", fmt.Errorf("file not found")
					}
					// Handle mkdir command
					if contains(cmd, "mkdir") {
						return "", nil
					}
				}
			}
			return "", nil
		},
	}

	cfg := &config.SMBConfig{
		ServerName:        "testserver",
		ServerIP:          "127.0.0.1",
		ShareName:         "testshare",
		Username:          "testuser",
		Password:          "testpass",
		Port:              445,
		AuthProtocol:      "ntlm",
		MaxRetries:        3,
		InitialRetryDelay: 0.01, // Short delay for testing
		MaxRetryDelay:     1.0,
		RetryBackoff:      2.0,
	}

	// Create a temporary test file
	tmpFile := "/tmp/test-retry-upload.txt"
	// Use os.WriteFile for actual file creation
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	err = UploadFile(tmpFile, "test/file.txt", cfg, false)

	if err != nil {
		t.Errorf("Expected upload to succeed after retries, got error: %v", err)
	}
	// Should be called at least 3 times (2 failures + 1 success)
	if callCount < 3 {
		t.Errorf("Expected at least 3 calls (for retries), got %d", callCount)
	}
}
