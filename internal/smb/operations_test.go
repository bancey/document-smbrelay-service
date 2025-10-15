package smb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

func TestUploadFile_InvalidServer(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "nonexistent-server",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err = UploadFile(tmpFile, "test/file.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

func TestUploadFile_MissingLocalFile(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock
	smbClientExec = NewMockExecutor()
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := UploadFile("/nonexistent/file.txt", "test/file.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading non-existent local file")
	}
	
	if err != nil && !contains(err.Error(), "not found") && !contains(err.Error(), "no such file") {
		t.Logf("Error message: %v", err)
	}
}

func TestUploadFile_RemotePathNormalization(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test with leading slash - should be normalized and succeed
	err = UploadFile(tmpFile, "/test/file.txt", cfg, false)

	if err != nil {
		t.Errorf("Expected successful upload with normalized path, got error: %v", err)
	}
}

func TestUploadFile_EmptyRemotePath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err = UploadFile(tmpFile, "", cfg, false)

	// Should succeed even with empty remote path
	if err != nil {
		t.Errorf("Expected successful upload with empty path, got error: %v", err)
	}
}

func TestUploadFile_NestedPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test with nested path
	err = UploadFile(tmpFile, "folder1/folder2/folder3/file.txt", cfg, false)

	if err != nil {
		t.Errorf("Expected successful upload with nested path, got error: %v", err)
	}
}

func TestUploadFile_OverwriteTrue(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test with overwrite enabled
	err = UploadFile(tmpFile, "test/file.txt", cfg, true)

	if err != nil {
		t.Errorf("Expected successful upload with overwrite=true, got error: %v", err)
	}
}

func TestUploadFile_OverwriteFalse(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates file exists
	callCount := 0
	mock := &MockSmbClientExecutor{
		ExecuteFunc: func(args []string) (string, error) {
			callCount++
			// First call checks if file exists (ls command)
			if callCount == 1 {
				// Return output that indicates file exists
				return "test/file.txt A 1024 Mon Jan  1 00:00:00 2024\n", nil
			}
			// Subsequent calls shouldn't happen
			return "", nil
		},
	}
	smbClientExec = mock
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test with overwrite disabled (should fail if file exists)
	err = UploadFile(tmpFile, "test/file.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when file exists and overwrite=false")
	}
	
	if err != nil && !contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestUploadFile_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}
	
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a larger temporary test file (1MB)
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-large-upload.txt")
	
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	
	err := os.WriteFile(tmpFile, largeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err = UploadFile(tmpFile, "test/largefile.bin", cfg, false)

	if err != nil {
		t.Errorf("Expected successful upload of large file, got error: %v", err)
	}
}

func TestUploadFile_SpecialCharactersInPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful upload
	smbClientExec = SetupSuccessfulMock()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test paths with spaces and other characters
	testPaths := []string{
		"test folder/file with spaces.txt",
		"test-folder/file-with-dashes.txt",
		"test_folder/file_with_underscores.txt",
	}

	for _, path := range testPaths {
		err = UploadFile(tmpFile, path, cfg, false)

		if err != nil {
			t.Errorf("Expected successful upload for path '%s', got error: %v", path, err)
		}
	}
}

func TestUploadFile_EmptyConfig(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock
	smbClientExec = NewMockExecutor()
	
	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName:   "",
		ServerIP:     "",
		ShareName:    "",
		Username:     "",
		Password:     "",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err = UploadFile(tmpFile, "test/file.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading with empty config")
	}
}

// Helper function for string contains check
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

// ============================================================================
// Extended Test Cases - Edge cases and additional scenarios
// ============================================================================

// Test upload with root directory path
func TestUploadFile_RootDirectory(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-root.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	// Test file in root directory (no subdirectories)
	err = UploadFile(tmpFile, "rootfile.txt", cfg, false)

	// Should fail due to invalid server, but test the path handling
	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with Windows-style backslashes
func TestUploadFile_BackslashPath(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-backslash.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	// Test with backslashes (Windows-style path)
	err = UploadFile(tmpFile, "folder\\subfolder\\file.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with very long path
func TestUploadFile_LongPath(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-long-path.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	// Create a very long path
	longPath := "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/file.txt"
	err = UploadFile(tmpFile, longPath, cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with single character filename
func TestUploadFile_SingleCharFilename(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "x.txt")
	err := os.WriteFile(tmpFile, []byte("x"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	err = UploadFile(tmpFile, "x", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with unicode characters in path
func TestUploadFile_UnicodePath(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-unicode.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	// Test with unicode characters
	err = UploadFile(tmpFile, "folder/文件.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with empty file
func TestUploadFile_EmptyFile(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-empty.txt")
	err := os.WriteFile(tmpFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	err = UploadFile(tmpFile, "empty.txt", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with different port numbers
func TestUploadFile_CustomPorts(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-port.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	ports := []int{139, 445, 1445, 8445}

	for _, port := range ports {
		t.Run("port_"+string(rune(port)), func(t *testing.T) {
			cfg := &config.SMBConfig{
				ServerName: "testserver",
				ServerIP:   "127.0.0.1",
				ShareName:  "testshare",
				Username:   "testuser",
				Password:   "testpass",
				Port:       port,
			}

			err = UploadFile(tmpFile, "test.txt", cfg, false)

			if err == nil {
				t.Error("Expected error when uploading to invalid server")
			}
		})
	}
}

// Test upload with various domain configurations
func TestUploadFile_DomainVariations(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-domain.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	domains := []string{"", "WORKGROUP", "DOMAIN", "corp.example.com"}

	for _, domain := range domains {
		t.Run("domain_"+domain, func(t *testing.T) {
			cfg := &config.SMBConfig{
				ServerName: "testserver",
				ServerIP:   "127.0.0.1",
				ShareName:  "testshare",
				Username:   "testuser",
				Password:   "testpass",
				Port:       445,
				Domain:     domain,
			}

			err = UploadFile(tmpFile, "test.txt", cfg, false)

			if err == nil {
				t.Error("Expected error when uploading to invalid server")
			}
		})
	}
}

// Test upload with file that has no extension
func TestUploadFile_NoExtension(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	err = UploadFile(tmpFile, "folder/noextension", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test upload with path that ends with slash
func TestUploadFile_PathEndsWithSlash(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-slash.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	err = UploadFile(tmpFile, "folder/", cfg, false)

	if err == nil {
		t.Error("Expected error when uploading to invalid server")
	}
}

// Test that connection errors are properly returned
func TestUploadFile_ConnectionErrorMessage(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-error.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := &config.SMBConfig{
		ServerName: "testserver",
		ServerIP:   "127.0.0.1",
		ShareName:  "testshare",
		Username:   "testuser",
		Password:   "testpass",
		Port:       445,
	}

	err = UploadFile(tmpFile, "test.txt", cfg, false)

	if err == nil {
		t.Error("Expected connection error")
	}

	if err != nil {
		errMsg := err.Error()
		// Should contain connection-related error message
		if !strings.Contains(errMsg, "connect") && !strings.Contains(errMsg, "refused") &&
			!strings.Contains(errMsg, "failed") && !strings.Contains(errMsg, "SMB") {
			t.Logf("Error message: %s", errMsg)
		}
	}
}
