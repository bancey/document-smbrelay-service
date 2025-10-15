package smb

import (
	"fmt"
	"strings"
)

// MockSmbClientExecutor is a mock implementation for testing
type MockSmbClientExecutor struct {
	// ExecuteFunc allows tests to define custom behavior
	ExecuteFunc func(args []string) (string, error)
	// LastArgs stores the last arguments passed to Execute
	LastArgs []string
	// CallCount tracks how many times Execute was called
	CallCount int
}

// Execute runs the mock function
func (m *MockSmbClientExecutor) Execute(args []string) (string, error) {
	m.LastArgs = args
	m.CallCount++
	
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(args)
	}
	
	// Default behavior: simulate connection refused
	return "", fmt.Errorf("smbclient command failed: exit status 1 (output: Connection to 127.0.0.1 failed)")
}

// NewMockExecutor creates a new mock executor with default behavior
func NewMockExecutor() *MockSmbClientExecutor {
	return &MockSmbClientExecutor{}
}

// SetupSuccessfulMock configures the mock to simulate successful operations
func SetupSuccessfulMock() *MockSmbClientExecutor {
	mock := &MockSmbClientExecutor{
		ExecuteFunc: func(args []string) (string, error) {
			// Check what command is being executed
			for i, arg := range args {
				if arg == "-c" && i+1 < len(args) {
					cmd := args[i+1]
					
					// Handle ls command
					if strings.HasPrefix(cmd, "ls") {
						// If it's just "ls" (health check), return directory listing
						if cmd == "ls" || cmd == "ls " || strings.TrimSpace(cmd) == "ls" {
							return "  .                                   D        0  Mon Jan  1 00:00:00 2024\n  ..                                  D        0  Mon Jan  1 00:00:00 2024\n\n\t\t64256 blocks of size 1024. 32128 blocks available\n", nil
						}
						// If it's "ls filename" (file existence check), return file not found
						return "NT_STATUS_NO_SUCH_FILE", fmt.Errorf("smbclient command failed: exit status 1")
					}
					
					// Handle put command (upload)
					if strings.Contains(cmd, "put") {
						return "putting file test.txt as test.txt (1.0 kb/s) (average 1.0 kb/s)\n", nil
					}
					
					// Handle mkdir command
					if strings.HasPrefix(cmd, "mkdir") {
						return "", nil
					}
				}
			}
			
			return "", nil
		},
	}
	return mock
}

// SetupFailureMock configures the mock to simulate failed operations
func SetupFailureMock(errorType string) *MockSmbClientExecutor {
	mock := &MockSmbClientExecutor{
		ExecuteFunc: func(args []string) (string, error) {
			switch errorType {
			case "connection_refused":
				return "Connection to 127.0.0.1 failed (Error NT_STATUS_CONNECTION_REFUSED)", fmt.Errorf("smbclient command failed: exit status 1")
			case "auth_failure":
				return "session setup failed: NT_STATUS_LOGON_FAILURE", fmt.Errorf("smbclient command failed: exit status 1")
			case "share_not_found":
				return "tree connect failed: NT_STATUS_BAD_NETWORK_NAME", fmt.Errorf("smbclient command failed: exit status 1")
			case "access_denied":
				return "NT_STATUS_ACCESS_DENIED", fmt.Errorf("smbclient command failed: exit status 1")
			case "file_exists":
				return "NT_STATUS_OBJECT_NAME_COLLISION", fmt.Errorf("smbclient command failed: exit status 1")
			default:
				return "Unknown error", fmt.Errorf("smbclient command failed: exit status 1")
			}
		},
	}
	return mock
}
