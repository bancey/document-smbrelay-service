package smb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// Test constants for SMB status codes
const (
	testStatusObjectNameNotFound = "NT_STATUS_OBJECT_NAME_NOT_FOUND"
	testStatusAccessDenied       = "NT_STATUS_ACCESS_DENIED"
	testStatusFileIsADirectory   = "NT_STATUS_FILE_IS_A_DIRECTORY"
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
		ExecuteFunc: func(_ []string) (string, error) {
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

func TestListFiles_Success(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock with custom output
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			output := `  .                                   D        0  Mon Jan  1 00:00:00 2024
  ..                                  D        0  Mon Jan  1 00:00:00 2024
  document.pdf                        A     1024  Mon Jan  1 12:34:56 2024
  folder1                             D        0  Mon Jan  1 10:00:00 2024
  report.docx                         A     2048  Mon Jan  1 14:22:33 2024

		65535 blocks of size 1024. 32768 blocks available`
			return output, nil
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	files, err := ListFiles("", cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// Check first file
	if files[0].Name != "document.pdf" {
		t.Errorf("Expected first file name to be 'document.pdf', got '%s'", files[0].Name)
	}
	if files[0].IsDir {
		t.Error("Expected first file to not be a directory")
	}
	if files[0].Size != 1024 {
		t.Errorf("Expected first file size to be 1024, got %d", files[0].Size)
	}

	// Check directory
	if files[1].Name != "folder1" {
		t.Errorf("Expected second file name to be 'folder1', got '%s'", files[1].Name)
	}
	if !files[1].IsDir {
		t.Error("Expected second file to be a directory")
	}

	// Verify the command that was executed
	if len(mockExec.LastArgs) == 0 {
		t.Fatal("Expected command to be executed")
	}
	// Should contain the ls command
	foundCmd := false
	for _, arg := range mockExec.LastArgs {
		if arg == "ls" {
			foundCmd = true
			break
		}
	}
	if !foundCmd {
		t.Error("Expected 'ls' command to be executed")
	}
}

func TestListFiles_WithPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			output := `  file1.txt                           A      512  Mon Jan  1 12:00:00 2024
  file2.txt                           A      256  Mon Jan  1 13:00:00 2024

		65535 blocks of size 1024. 32768 blocks available`
			return output, nil
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	files, err := ListFiles("subfolder", cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Verify the command includes cd
	if len(mockExec.LastArgs) == 0 {
		t.Fatal("Expected command to be executed")
	}
	foundCd := false
	for _, arg := range mockExec.LastArgs {
		if arg == `cd "subfolder"; ls` {
			foundCd = true
			break
		}
	}
	if !foundCd {
		t.Error("Expected 'cd' command to be executed")
	}
}

func TestListFiles_PathNotFound(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return testStatusObjectNameNotFound, fmt.Errorf("smbclient command failed")
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	_, err := ListFiles("nonexistent", cfg)
	if err == nil {
		t.Fatal("Expected error for non-existent path")
	}

	if err.Error() != "path not found: nonexistent" {
		t.Errorf("Expected 'path not found' error, got: %v", err)
	}
}

func TestListFiles_AccessDenied(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return testStatusAccessDenied, fmt.Errorf("smbclient command failed")
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	_, err := ListFiles("protected", cfg)
	if err == nil {
		t.Fatal("Expected error for access denied")
	}

	if err.Error() != "access denied to path: protected" {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestParseLsOutput(t *testing.T) {
	//nolint:govet // fieldalignment: test struct readability is more important than memory optimization
	tests := []struct {
		name     string
		output   string
		expected int
		validate func(*testing.T, []FileInfo)
	}{
		{
			name: "Standard output with files and folders",
			output: `  .                                   D        0  Mon Jan  1 00:00:00 2024
  ..                                  D        0  Mon Jan  1 00:00:00 2024
  file1.txt                           A      512  Mon Jan  1 12:00:00 2024
  folder1                             D        0  Mon Jan  1 10:00:00 2024
  file2.pdf                           A     1024  Mon Jan  1 14:00:00 2024

		65535 blocks of size 1024. 32768 blocks available`,
			expected: 3, // . and .. should be filtered out
		},
		{
			name: "Empty directory",
			output: `  .                                   D        0  Mon Jan  1 00:00:00 2024
  ..                                  D        0  Mon Jan  1 00:00:00 2024

		65535 blocks of size 1024. 32768 blocks available`,
			expected: 0,
		},
		{
			name:     "Invalid output",
			output:   "Some error message",
			expected: 0,
		},
		{
			name: "Filenames with spaces",
			output: `  .                                   D        0  Mon Jan  1 00:00:00 2024
  ..                                  D        0  Mon Jan  1 00:00:00 2024
  Quarterly Report.pdf                A     2048  Mon Jan  1 12:34:56 2024
  My Documents                        D        0  Mon Jan  1 10:00:00 2024
  Sales Data 2024.xlsx                A     4096  Mon Jan  1 14:22:33 2024
  Project Files                       D        0  Mon Jan  1 09:15:00 2024

		65535 blocks of size 1024. 32768 blocks available`,
			expected: 4,
			validate: func(t *testing.T, files []FileInfo) {
				// Verify that filenames with spaces are preserved
				expectedNames := map[string]bool{
					"Quarterly Report.pdf": false,
					"My Documents":         false,
					"Sales Data 2024.xlsx": false,
					"Project Files":        false,
				}

				for _, file := range files {
					if _, exists := expectedNames[file.Name]; exists {
						expectedNames[file.Name] = true
					}
				}

				for name, found := range expectedNames {
					if !found {
						t.Errorf("Expected to find file '%s' but it was not parsed", name)
					}
				}

				// Verify specific file attributes
				for _, file := range files {
					if file.Name == "Quarterly Report.pdf" {
						if file.IsDir {
							t.Errorf("'Quarterly Report.pdf' should not be a directory")
						}
						if file.Size != 2048 {
							t.Errorf("'Quarterly Report.pdf' should have size 2048, got %d", file.Size)
						}
					}
					if file.Name == "My Documents" {
						if !file.IsDir {
							t.Errorf("'My Documents' should be a directory")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := parseLsOutput(tt.output)
			if len(files) != tt.expected {
				t.Errorf("Expected %d files, got %d", tt.expected, len(files))
				for i, f := range files {
					t.Logf("File %d: Name='%s', IsDir=%v, Size=%d", i, f.Name, f.IsDir, f.Size)
				}
			}

			if tt.validate != nil {
				tt.validate(t, files)
			}
		})
	}
}

func TestListFiles_NormalizePath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	tests := []struct {
		name       string
		inputPath  string
		expectPath string
	}{
		{
			name:       "Forward slash prefix",
			inputPath:  "/folder",
			expectPath: "folder",
		},
		{
			name:       "Backslash prefix",
			inputPath:  "\\folder",
			expectPath: "folder",
		},
		{
			name:       "Mixed slashes",
			inputPath:  "folder\\subfolder",
			expectPath: "folder/subfolder",
		},
		{
			name:       "No prefix",
			inputPath:  "folder",
			expectPath: "folder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &MockSmbClientExecutor{
				ExecuteFunc: func(_ []string) (string, error) {
					return "", nil
				},
			}
			smbClientExec = mockExec

			cfg := &config.SMBConfig{
				ServerName:   "testserver",
				ServerIP:     "192.168.1.100",
				ShareName:    "testshare",
				Username:     "testuser",
				Password:     "testpass",
				Port:         445,
				AuthProtocol: "ntlm",
			}

			_, _ = ListFiles(tt.inputPath, cfg)

			// Check that the path was normalized in the command
			if len(mockExec.LastArgs) == 0 {
				t.Fatal("Expected command to be executed")
			}
			// The normalized path should be in the command
			foundCmd := false
			for _, arg := range mockExec.LastArgs {
				if tt.expectPath != "" && tt.expectPath != "." {
					expectedCmd := fmt.Sprintf(`cd "%s"; ls`, tt.expectPath)
					if arg == expectedCmd {
						foundCmd = true
						break
					}
				} else if arg == "ls" {
					foundCmd = true
					break
				}
			}
			if !foundCmd {
				t.Errorf("Expected normalized path in command, got args: %v", mockExec.LastArgs)
			}
		})
	}
}

// ============================================================================
// Delete File Tests
// ============================================================================

func TestDeleteFile_Success(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that simulates successful delete
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return "deleted file successfully", nil
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("folder/file.txt", cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the command that was executed
	if len(mockExec.LastArgs) == 0 {
		t.Fatal("Expected command to be executed")
	}
	foundCmd := false
	for _, arg := range mockExec.LastArgs {
		if strings.Contains(arg, "del") && strings.Contains(arg, "folder/file.txt") {
			foundCmd = true
			break
		}
	}
	if !foundCmd {
		t.Error("Expected 'del' command to be executed")
	}
}

func TestDeleteFile_FileNotFound(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that simulates file not found
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return testStatusObjectNameNotFound, fmt.Errorf("smbclient command failed")
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("nonexistent.txt", cfg)
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}
}

func TestDeleteFile_AccessDenied(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that simulates access denied
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return testStatusAccessDenied, fmt.Errorf("smbclient command failed")
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("protected.txt", cfg)
	if err == nil {
		t.Fatal("Expected error for access denied")
	}

	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestDeleteFile_IsDirectory(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that simulates attempting to delete a directory
	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return testStatusFileIsADirectory, fmt.Errorf("smbclient command failed")
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("folder", cfg)
	if err == nil {
		t.Fatal("Expected error when deleting directory")
	}

	if !strings.Contains(err.Error(), "cannot delete directory") {
		t.Errorf("Expected 'cannot delete directory' error, got: %v", err)
	}
}

func TestDeleteFile_EmptyPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	mockExec := NewMockExecutor()
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("", cfg)
	if err == nil {
		t.Fatal("Expected error for empty path")
	}

	if !strings.Contains(err.Error(), "invalid remote path") {
		t.Errorf("Expected 'invalid remote path' error, got: %v", err)
	}
}

func TestDeleteFile_RootPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	mockExec := NewMockExecutor()
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test various root path formats
	testPaths := []string{".", "/", "\\"}
	for _, path := range testPaths {
		err := DeleteFile(path, cfg)
		if err == nil {
			t.Errorf("Expected error for root path '%s'", path)
		}
		if !strings.Contains(err.Error(), "invalid remote path") && !strings.Contains(err.Error(), "cannot delete root") {
			t.Errorf("Expected 'invalid remote path' error for '%s', got: %v", path, err)
		}
	}
}

func TestDeleteFile_PathNormalization(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	tests := []struct {
		name       string
		inputPath  string
		expectPath string
	}{
		{
			name:       "Forward slash prefix",
			inputPath:  "/folder/file.txt",
			expectPath: "folder/file.txt",
		},
		{
			name:       "Backslash prefix",
			inputPath:  "\\folder\\file.txt",
			expectPath: "folder/file.txt",
		},
		{
			name:       "Mixed slashes",
			inputPath:  "folder\\subfolder/file.txt",
			expectPath: "folder/subfolder/file.txt",
		},
		{
			name:       "No prefix",
			inputPath:  "folder/file.txt",
			expectPath: "folder/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &MockSmbClientExecutor{
				ExecuteFunc: func(_ []string) (string, error) {
					return "deleted", nil
				},
			}
			smbClientExec = mockExec

			cfg := &config.SMBConfig{
				ServerName:   "testserver",
				ServerIP:     "192.168.1.100",
				ShareName:    "testshare",
				Username:     "testuser",
				Password:     "testpass",
				Port:         445,
				AuthProtocol: "ntlm",
			}

			err := DeleteFile(tt.inputPath, cfg)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			// Check that the path was normalized in the command
			if len(mockExec.LastArgs) == 0 {
				t.Fatal("Expected command to be executed")
			}
			foundCmd := false
			for _, arg := range mockExec.LastArgs {
				expectedCmd := fmt.Sprintf(`del "%s"`, tt.expectPath)
				if arg == expectedCmd {
					foundCmd = true
					break
				}
			}
			if !foundCmd {
				t.Errorf("Expected normalized path in command, got args: %v", mockExec.LastArgs)
			}
		})
	}
}

func TestDeleteFile_SpecialCharactersInPath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	mockExec := &MockSmbClientExecutor{
		ExecuteFunc: func(_ []string) (string, error) {
			return "deleted", nil
		},
	}
	smbClientExec = mockExec

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Test paths with special characters
	testPaths := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"folder/file with spaces.txt",
	}

	for _, path := range testPaths {
		err := DeleteFile(path, cfg)
		if err != nil {
			t.Errorf("Expected successful delete for path '%s', got error: %v", path, err)
		}
	}
}

func TestDeleteFile_ConnectionError(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	err := DeleteFile("test.txt", cfg)
	if err == nil {
		t.Error("Expected error when deleting from invalid server")
	}
}

// ============================================================================
// Path Utility Function Tests
// ============================================================================

func TestNormalizePathSegment(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Single dot", ".", "."},
		{"Forward slash", "/", ""},
		{"Backslash", "\\", ""},
		{"Leading slash", "/path/to/file", "path/to/file"},
		{"Trailing slash", "path/to/file/", "path/to/file"},
		{"Both slashes", "/path/to/file/", "path/to/file"},
		{"Backslashes to forward", "path\\to\\file", "path/to/file"},
		{"Mixed slashes", "/path\\to/file\\", "path/to/file"},
		{"Double slashes", "path//to//file", "path/to/file"},
		{"Multiple leading slashes", "///path/to/file", "path/to/file"},
		{"Multiple trailing slashes", "path/to/file///", "path/to/file"},
		{"Complex path", "\\\\path//to\\\\file//", "path/to/file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePathSegment(tc.input)
			if result != tc.expected {
				t.Errorf("For input '%s', expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		})
	}
}

func TestJoinSmbPaths(t *testing.T) {
	testCases := []struct {
		name         string
		basePath     string
		relativePath string
		expected     string
	}{
		{"Both empty", "", "", ""},
		{"Base empty, relative set", "", "file.txt", "file.txt"},
		{"Base set, relative empty", "apps/myapp", "", "apps/myapp"},
		{"Both set simple", "apps/myapp", "file.txt", "apps/myapp/file.txt"},
		{"Both set with subdirs", "apps/myapp", "inbox/file.txt", "apps/myapp/inbox/file.txt"},
		{"Base with trailing slash", "apps/myapp/", "file.txt", "apps/myapp/file.txt"},
		{"Relative with leading slash", "apps/myapp", "/file.txt", "apps/myapp/file.txt"},
		{"Both with slashes", "apps/myapp/", "/file.txt", "apps/myapp/file.txt"},
		{"Base with backslashes", "apps\\myapp", "file.txt", "apps/myapp/file.txt"},
		{"Relative with backslashes", "apps/myapp", "inbox\\file.txt", "apps/myapp/inbox/file.txt"},
		{"Mixed slashes", "apps\\myapp/", "\\inbox/file.txt", "apps/myapp/inbox/file.txt"},
		{"Base is dot", ".", "file.txt", "file.txt"},
		{"Relative is dot", "apps/myapp", ".", "apps/myapp"},
		{"Both are dots", ".", ".", "."},
		{"Complex path", "/apps//myapp\\", "\\inbox//file.txt", "apps/myapp/inbox/file.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := joinSmbPaths(tc.basePath, tc.relativePath)
			if result != tc.expected {
				t.Errorf("joinSmbPaths('%s', '%s'): expected '%s', got '%s'",
					tc.basePath, tc.relativePath, tc.expected, result)
			}
		})
	}
}

func TestBuildFullPath(t *testing.T) {
	testCases := []struct {
		name         string
		basePath     string
		relativePath string
		expected     string
	}{
		{"No base path", "", "file.txt", "file.txt"},
		{"Simple base path", "apps/myapp", "file.txt", "apps/myapp/file.txt"},
		{"Nested relative path", "apps/myapp", "inbox/document.pdf", "apps/myapp/inbox/document.pdf"},
		{"Base with trailing slash", "apps/myapp/", "file.txt", "apps/myapp/file.txt"},
		{"Relative with leading slash", "apps/myapp", "/file.txt", "apps/myapp/file.txt"},
		{"Backslash normalization", "apps\\myapp", "inbox\\file.txt", "apps/myapp/inbox/file.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.SMBConfig{
				BasePath: tc.basePath,
			}
			result := buildFullPath(tc.relativePath, cfg)
			if result != tc.expected {
				t.Errorf("buildFullPath('%s', cfg{BasePath:'%s'}): expected '%s', got '%s'",
					tc.relativePath, tc.basePath, tc.expected, result)
			}
		})
	}
}

// ============================================================================
// Integration Tests with Base Path
// ============================================================================

func TestListFiles_WithBasePath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Use mock that returns a simple file listing
	mockOutput := `  file1.txt                           A     1024  Mon Jan  1 12:00:00 2024
  file2.txt                           A     2048  Mon Jan  1 12:00:00 2024

		32768 blocks of size 1024. 16384 blocks available`
	smbClientExec = NewMockExecutorWithOutput(mockOutput)

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "data",
		BasePath:     "apps/myapp",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// List files in "inbox" which should resolve to "apps/myapp/inbox"
	files, err := ListFiles("inbox", cfg)
	if err != nil {
		t.Fatalf("Expected successful listing, got error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestUploadFile_WithBasePath(t *testing.T) {
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
		ShareName:    "data",
		BasePath:     "apps/myapp",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Upload to "inbox/file.txt" which should resolve to "apps/myapp/inbox/file.txt"
	err = UploadFile(tmpFile, "inbox/file.txt", cfg, false)
	if err != nil {
		t.Errorf("Expected successful upload, got error: %v", err)
	}
}

func TestDeleteFile_WithBasePath(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Use mock that simulates successful deletion
	smbClientExec = SetupSuccessfulMock()

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "data",
		BasePath:     "apps/myapp",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	// Delete "inbox/file.txt" which should resolve to "apps/myapp/inbox/file.txt"
	err := DeleteFile("inbox/file.txt", cfg)
	if err != nil {
		t.Errorf("Expected successful deletion, got error: %v", err)
	}
}
