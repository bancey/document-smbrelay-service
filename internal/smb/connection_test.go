package smb

import (
	"os"
	"testing"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

func TestCheckHealth_MissingConfig(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	// Clear environment
	os.Clearenv()

	cfg := &config.SMBConfig{
		ServerName: "",
		ServerIP:   "",
		ShareName:  "",
		Username:   "",
		Password:   "",
		Port:       445,
	}

	result := CheckHealth(cfg)

	if result.AppStatus != "ok" {
		t.Errorf("Expected app_status to be 'ok', got '%s'", result.AppStatus)
	}

	if result.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", result.Status)
	}

	if result.SMBConnection != "failed" {
		t.Errorf("Expected smb_connection to be 'failed', got '%s'", result.SMBConnection)
	}

	if result.SMBShareAccessible {
		t.Error("Expected smb_share_accessible to be false")
	}

	if result.Error == "" {
		t.Error("Expected error message to be present")
	}
}

func TestCheckHealth_InvalidServer(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "nonexistent-server",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.AppStatus != "ok" {
		t.Errorf("Expected app_status to be 'ok', got '%s'", result.AppStatus)
	}

	if result.Status != "unhealthy" {
		t.Errorf("Expected status to be 'unhealthy', got '%s'", result.Status)
	}

	if result.SMBConnection != "failed" {
		t.Errorf("Expected smb_connection to be 'failed', got '%s'", result.SMBConnection)
	}

	if result.SMBShareAccessible {
		t.Error("Expected smb_share_accessible to be false")
	}

	if result.Error == "" {
		t.Error("Expected error message to be present")
	}
}

func TestHealthCheckResult_Fields(t *testing.T) {
	result := &HealthCheckResult{
		Status:             "healthy",
		AppStatus:          "ok",
		SMBConnection:      "ok",
		SMBShareAccessible: true,
		Server:             "testserver (127.0.0.1:445)",
		Share:              "testshare",
		Error:              "",
	}

	if result.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", result.Status)
	}

	if result.AppStatus != "ok" {
		t.Errorf("Expected app_status 'ok', got '%s'", result.AppStatus)
	}

	if result.SMBConnection != "ok" {
		t.Errorf("Expected smb_connection 'ok', got '%s'", result.SMBConnection)
	}

	if !result.SMBShareAccessible {
		t.Error("Expected smb_share_accessible to be true")
	}

	if result.Server != "testserver (127.0.0.1:445)" {
		t.Errorf("Expected server 'testserver (127.0.0.1:445)', got '%s'", result.Server)
	}

	if result.Share != "testshare" {
		t.Errorf("Expected share 'testshare', got '%s'", result.Share)
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got '%s'", result.Error)
	}
}

func TestCheckHealth_ConnectionRefused(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "nonexistent",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         9999, // Invalid port
		Domain:       "",
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}

	if result.Error == "" {
		t.Error("Expected error message to be present")
	}
}

func TestCheckHealth_EmptyCredentials(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock
	smbClientExec = NewMockExecutor()
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "",
		Password:     "",
		Port:         445,
		Domain:       "",
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	// Should fail due to missing credentials
	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}
	
	if result.Error == "" {
		t.Error("Expected error message for missing credentials")
	}
}

func TestCheckHealth_ServerDisplay(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.10",
		ShareName:    "documents",
		Username:     "user",
		Password:     "pass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	expectedServer := "testserver (192.168.1.10:445)"
	if result.Server != expectedServer {
		t.Errorf("Expected server '%s', got '%s'", expectedServer, result.Server)
	}

	expectedShare := "documents"
	if result.Share != expectedShare {
		t.Errorf("Expected share '%s', got '%s'", expectedShare, result.Share)
	}
}

func TestCheckHealth_CustomPort(t *testing.T) {
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
		Port:         1445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	// Should fail because server doesn't exist
	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}

	// Check that the port is included in server display
	if result.Server != "testserver (127.0.0.1:1445)" {
		t.Errorf("Expected server to include custom port 1445, got '%s'", result.Server)
	}
}

func TestCheckHealth_SuccessfulConnection(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful connection
	smbClientExec = SetupSuccessfulMock()
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "testshare",
		Username:     "user",
		Password:     "pass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", result.Status)
	}

	if result.SMBConnection != "ok" {
		t.Errorf("Expected smb_connection 'ok', got '%s'", result.SMBConnection)
	}

	if !result.SMBShareAccessible {
		t.Error("Expected smb_share_accessible to be true")
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got '%s'", result.Error)
	}
}

// ============================================================================
// Extended Test Cases - Edge cases and additional scenarios
// ============================================================================

// Test CheckHealth with different port configurations
func TestCheckHealth_VariousPorts(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	ports := []int{139, 445, 1445, 8445}

	for _, port := range ports {
		cfg := &config.SMBConfig{
			ServerName:   "testserver",
			ServerIP:     "127.0.0.1",
			ShareName:    "testshare",
			Username:     "testuser",
			Password:     "testpass",
			Port:         port,
			AuthProtocol: "ntlm",
		}

		result := CheckHealth(cfg)

		if result.Status != "unhealthy" {
			t.Errorf("Port %d: Expected status 'unhealthy', got '%s'", port, result.Status)
		}

		if result.AppStatus != "ok" {
			t.Errorf("Port %d: Expected app_status 'ok', got '%s'", port, result.AppStatus)
		}
	}
}

// Test CheckHealth with IPv6 address
func TestCheckHealth_IPv6Address(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "::1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}
}

// Test CheckHealth with domain
func TestCheckHealth_WithDomain(t *testing.T) {
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
		Domain:       "WORKGROUP",
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}
}

// Test CheckHealth with empty share name
func TestCheckHealth_EmptyShareName(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}
}

// Test CheckHealth with various credentials
func TestCheckHealth_Credentials(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	testCases := []struct {
		name     string
		username string
		password string
		authType string
	}{
		{"valid_creds", "user", "pass", "ntlm"},
		{"empty_username", "", "pass", "ntlm"},
		{"empty_password", "user", "", "ntlm"},
		{"kerberos", "", "", "kerberos"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use mock that simulates connection failure
			smbClientExec = SetupFailureMock("connection_refused")
			
			cfg := &config.SMBConfig{
				ServerName:   "testserver",
				ServerIP:     "127.0.0.1",
				ShareName:    "testshare",
				Username:     tc.username,
				Password:     tc.password,
				Port:         445,
				AuthProtocol: tc.authType,
			}

			result := CheckHealth(cfg)

			// All should be unhealthy due to connection failure or missing creds
			if result.Status != "unhealthy" {
				t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
			}
		})
	}
}

// Test CheckHealth server and share info display
func TestCheckHealth_ServerShareInfo(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "fileserver",
		ServerIP:     "192.168.1.100",
		ShareName:    "documents",
		Username:     "user",
		Password:     "pass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	expectedServer := "fileserver (192.168.1.100:445)"
	if result.Server != expectedServer {
		t.Errorf("Expected server '%s', got '%s'", expectedServer, result.Server)
	}

	if result.Share != "documents" {
		t.Errorf("Expected share 'documents', got '%s'", result.Share)
	}
}

// Test CheckHealth with hostname only (no IP)
func TestCheckHealth_HostnameOnly(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates connection failure
	smbClientExec = SetupFailureMock("connection_refused")
	
	cfg := &config.SMBConfig{
		ServerName:   "fileserver",
		ServerIP:     "",
		ShareName:    "share",
		Username:     "user",
		Password:     "pass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if result.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", result.Status)
	}
}

// Test that error field is populated on failure
func TestCheckHealth_ErrorField(t *testing.T) {
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

	result := CheckHealth(cfg)

	if result.Error == "" {
		t.Error("Expected error field to be populated on failure")
	}
}

// Test SMB connection with different authentication protocols
func TestCheckHealth_AuthProtocols(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	protocols := []string{"ntlm", "negotiate", "kerberos"}

	for _, protocol := range protocols {
		t.Run(protocol, func(t *testing.T) {
			// Use mock that simulates connection failure
			smbClientExec = SetupFailureMock("connection_refused")
			
			cfg := &config.SMBConfig{
				ServerName:   "testserver",
				ServerIP:     "127.0.0.1",
				ShareName:    "testshare",
				Username:     "testuser",
				Password:     "testpass",
				Port:         445,
				AuthProtocol: protocol,
			}

			result := CheckHealth(cfg)

			// Should be unhealthy due to connection failure
			if result.Status != "unhealthy" {
				t.Errorf("Protocol %s: Expected status 'unhealthy', got '%s'", protocol, result.Status)
			}
		})
	}
}

// Test CheckHealth with successful connection
func TestCheckHealth_ShareAccessibleDefault(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	// Use mock that simulates successful connection
	smbClientExec = SetupSuccessfulMock()
	
	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		Port:         445,
		AuthProtocol: "ntlm",
	}

	result := CheckHealth(cfg)

	if !result.SMBShareAccessible {
		t.Error("Expected smb_share_accessible to be true")
	}
}

// Test that error contains meaningful details
func TestCheckHealth_ErrorContainsDetail(t *testing.T) {
	// Save original executor and restore after test
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()
	
	testCases := []struct {
		name          string
		mockType      string
		expectedError string
	}{
		{"connection_refused", "connection_refused", "connection refused"},
		{"auth_failure", "auth_failure", "authentication failed"},
		{"share_not_found", "share_not_found", "share not found"},
		{"access_denied", "access_denied", "access denied"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use specific failure mock
			smbClientExec = SetupFailureMock(tc.mockType)
			
			cfg := &config.SMBConfig{
				ServerName:   "testserver",
				ServerIP:     "127.0.0.1",
				ShareName:    "testshare",
				Username:     "testuser",
				Password:     "testpass",
				Port:         445,
				AuthProtocol: "ntlm",
			}

			result := CheckHealth(cfg)

			if result.Error == "" {
				t.Error("Expected error message to be present")
			}
		})
	}
}
