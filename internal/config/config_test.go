package config

import (
	"os"
	"testing"
)

const (
	testAuthProtocolNTLM      = "ntlm"
	testAuthProtocolNegotiate = "negotiate"
)

func TestLoadFromEnv(t *testing.T) {
	// Test with missing required variables
	t.Run("Missing required variables", func(t *testing.T) {
		os.Clearenv()
		_, missing := LoadFromEnv()

		if len(missing) == 0 {
			t.Error("Expected missing variables, got none")
		}

		expectedMissing := []string{"SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD"}
		if len(missing) != len(expectedMissing) {
			t.Errorf("Expected %d missing variables, got %d", len(expectedMissing), len(missing))
		}
	})

	// Test with all required variables
	t.Run("All required variables present", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")

		cfg, missing := LoadFromEnv()

		if len(missing) != 0 {
			t.Errorf("Expected no missing variables, got %v", missing)
		}

		if cfg.ServerName != "testserver" {
			t.Errorf("Expected ServerName 'testserver', got '%s'", cfg.ServerName)
		}

		if cfg.Port != 445 {
			t.Errorf("Expected Port 445, got %d", cfg.Port)
		}

		if cfg.AuthProtocol != testAuthProtocolNTLM {
			t.Errorf("Expected AuthProtocol 'ntlm', got '%s'", cfg.AuthProtocol)
		}
	})

	// Test custom port
	t.Run("Custom port", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_PORT", "1445")

		cfg, _ := LoadFromEnv()

		if cfg.Port != 1445 {
			t.Errorf("Expected Port 1445, got %d", cfg.Port)
		}
	})

	// Test invalid port
	t.Run("Invalid port defaults to 445", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_PORT", "invalid")

		cfg, _ := LoadFromEnv()

		if cfg.Port != 445 {
			t.Errorf("Expected Port 445 for invalid port, got %d", cfg.Port)
		}
	})

	// Test domain configuration
	t.Run("Domain configuration", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_DOMAIN", "WORKGROUP")

		cfg, _ := LoadFromEnv()

		if cfg.Domain != "WORKGROUP" {
			t.Errorf("Expected Domain 'WORKGROUP', got '%s'", cfg.Domain)
		}
	})

	// Test UseNTLMv2 variations
	t.Run("UseNTLMv2 true variations", func(t *testing.T) {
		testCases := []string{"true", "TRUE", "True", "1", "yes", "YES"}

		for _, val := range testCases {
			os.Clearenv()
			os.Setenv("SMB_SERVER_NAME", "testserver")
			os.Setenv("SMB_SERVER_IP", "127.0.0.1")
			os.Setenv("SMB_SHARE_NAME", "testshare")
			os.Setenv("SMB_USERNAME", "testuser")
			os.Setenv("SMB_PASSWORD", "testpass")
			os.Setenv("SMB_USE_NTLM_V2", val)

			cfg, _ := LoadFromEnv()

			if !cfg.UseNTLMv2 {
				t.Errorf("Expected UseNTLMv2 to be true for value '%s'", val)
			}
		}
	})

	// Test UseNTLMv2 false
	t.Run("UseNTLMv2 false", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_USE_NTLM_V2", "false")

		cfg, _ := LoadFromEnv()

		if cfg.UseNTLMv2 {
			t.Error("Expected UseNTLMv2 to be false")
		}
	})

	// Test AuthProtocol explicit setting
	t.Run("AuthProtocol negotiate", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_AUTH_PROTOCOL", "negotiate")

		cfg, _ := LoadFromEnv()

		if cfg.AuthProtocol != testAuthProtocolNegotiate {
			t.Errorf("Expected AuthProtocol 'negotiate', got '%s'", cfg.AuthProtocol)
		}
	})

	// Test AuthProtocol ntlm
	t.Run("AuthProtocol ntlm", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_AUTH_PROTOCOL", "ntlm")

		cfg, _ := LoadFromEnv()

		if cfg.AuthProtocol != testAuthProtocolNTLM {
			t.Errorf("Expected AuthProtocol 'ntlm', got '%s'", cfg.AuthProtocol)
		}
	})

	// Test AuthProtocol invalid defaults to derive from UseNTLMv2
	t.Run("AuthProtocol invalid", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_AUTH_PROTOCOL", "invalid")
		os.Setenv("SMB_USE_NTLM_V2", "true")

		cfg, _ := LoadFromEnv()

		if cfg.AuthProtocol != testAuthProtocolNTLM {
			t.Errorf("Expected AuthProtocol to default to 'ntlm' when invalid, got '%s'", cfg.AuthProtocol)
		}
	})

	// Test backward compatibility: UseNTLMv2 derives AuthProtocol
	t.Run("Backward compatibility UseNTLMv2 true", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_USE_NTLM_V2", "true")

		cfg, _ := LoadFromEnv()

		if cfg.AuthProtocol != testAuthProtocolNTLM {
			t.Errorf("Expected AuthProtocol 'ntlm' from UseNTLMv2=true, got '%s'", cfg.AuthProtocol)
		}
	})

	// Test backward compatibility: UseNTLMv2 false derives AuthProtocol
	t.Run("Backward compatibility UseNTLMv2 false", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMB_SERVER_NAME", "testserver")
		os.Setenv("SMB_SERVER_IP", "127.0.0.1")
		os.Setenv("SMB_SHARE_NAME", "testshare")
		os.Setenv("SMB_USERNAME", "testuser")
		os.Setenv("SMB_PASSWORD", "testpass")
		os.Setenv("SMB_USE_NTLM_V2", "false")

		cfg, _ := LoadFromEnv()

		if cfg.AuthProtocol != testAuthProtocolNegotiate {
			t.Errorf("Expected AuthProtocol 'negotiate' from UseNTLMv2=false, got '%s'", cfg.AuthProtocol)
		}
	})
}

func TestGetServer(t *testing.T) {
	cfg := &SMBConfig{
		ServerName: "testserver",
		ServerIP:   "192.168.1.10",
		Port:       445,
	}

	server := cfg.GetServer()
	expected := "192.168.1.10:445"

	if server != expected {
		t.Errorf("Expected '%s', got '%s'", expected, server)
	}
}

func TestGetServerDisplay(t *testing.T) {
	cfg := &SMBConfig{
		ServerName: "testserver",
		ServerIP:   "192.168.1.10",
		Port:       445,
	}

	display := cfg.GetServerDisplay()
	expected := "testserver (192.168.1.10:445)"

	if display != expected {
		t.Errorf("Expected '%s', got '%s'", expected, display)
	}
}

// ============================================================================
// Extended Test Cases - Edge cases and additional scenarios
// ============================================================================

// Test edge case where server IP is empty but server name is provided
func TestGetServer_EmptyIP(t *testing.T) {
	cfg := &SMBConfig{
		ServerName: "myserver",
		ServerIP:   "",
		Port:       445,
	}

	server := cfg.GetServer()
	expected := "myserver:445"

	if server != expected {
		t.Errorf("Expected '%s', got '%s'", expected, server)
	}
}

// Test edge case where both server name and IP are empty
func TestGetServer_BothEmpty(t *testing.T) {
	cfg := &SMBConfig{
		ServerName: "",
		ServerIP:   "",
		Port:       445,
	}

	server := cfg.GetServer()
	expected := ":445"

	if server != expected {
		t.Errorf("Expected '%s', got '%s'", expected, server)
	}
}

// Test all combinations of missing variables
func TestLoadFromEnv_PartiallyMissing(t *testing.T) {
	tests := []struct {
		envVars       map[string]string
		name          string
		expectedCount int
	}{
		{
			name: "Only server name missing",
			envVars: map[string]string{
				"SMB_SERVER_IP":  "127.0.0.1",
				"SMB_SHARE_NAME": "share",
				"SMB_USERNAME":   "user",
				"SMB_PASSWORD":   "pass",
			},
			expectedCount: 1,
		},
		{
			name: "Only server IP missing",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "server",
				"SMB_SHARE_NAME":  "share",
				"SMB_USERNAME":    "user",
				"SMB_PASSWORD":    "pass",
			},
			expectedCount: 1,
		},
		{
			name: "Only share name missing",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "server",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_USERNAME":    "user",
				"SMB_PASSWORD":    "pass",
			},
			expectedCount: 1,
		},
		{
			name: "Only username missing",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "server",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_SHARE_NAME":  "share",
				"SMB_PASSWORD":    "pass",
			},
			expectedCount: 1,
		},
		{
			name: "Only password missing",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "server",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_SHARE_NAME":  "share",
				"SMB_USERNAME":    "user",
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			_, missing := LoadFromEnv()

			if len(missing) != tt.expectedCount {
				t.Errorf("Expected %d missing variables, got %d: %v", tt.expectedCount, len(missing), missing)
			}
		})
	}
}

// Test all boolean variations for UseNTLMv2
func TestLoadFromEnv_BooleanParsing(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"lowercase true", "true", true},
		{"uppercase TRUE", "TRUE", true},
		{"mixed case True", "True", true},
		{"numeric 1", "1", true},
		{"lowercase yes", "yes", true},
		{"uppercase YES", "YES", true},
		{"lowercase false", "false", false},
		{"uppercase FALSE", "FALSE", false},
		{"numeric 0", "0", false},
		{"lowercase no", "no", false},
		{"random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("SMB_SERVER_NAME", "test")
			os.Setenv("SMB_SERVER_IP", "127.0.0.1")
			os.Setenv("SMB_SHARE_NAME", "share")
			os.Setenv("SMB_USERNAME", "user")
			os.Setenv("SMB_PASSWORD", "pass")
			os.Setenv("SMB_USE_NTLM_V2", tt.value)

			cfg, _ := LoadFromEnv()

			if cfg.UseNTLMv2 != tt.expected {
				t.Errorf("For value '%s', expected UseNTLMv2=%v, got %v", tt.value, tt.expected, cfg.UseNTLMv2)
			}
		})
	}
}

// Test empty port string defaults to 445
func TestLoadFromEnv_EmptyPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "test")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "share")
	os.Setenv("SMB_USERNAME", "user")
	os.Setenv("SMB_PASSWORD", "pass")
	os.Setenv("SMB_PORT", "")

	cfg, _ := LoadFromEnv()

	if cfg.Port != 445 {
		t.Errorf("Expected default port 445, got %d", cfg.Port)
	}
}

// Test negative port number
func TestLoadFromEnv_NegativePort(t *testing.T) {
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "test")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "share")
	os.Setenv("SMB_USERNAME", "user")
	os.Setenv("SMB_PASSWORD", "pass")
	os.Setenv("SMB_PORT", "-445")

	cfg, _ := LoadFromEnv()

	if cfg.Port != -445 {
		t.Errorf("Expected port -445, got %d", cfg.Port)
	}
}

// Test port with spaces
func TestLoadFromEnv_PortWithSpaces(t *testing.T) {
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "test")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "share")
	os.Setenv("SMB_USERNAME", "user")
	os.Setenv("SMB_PASSWORD", "pass")
	os.Setenv("SMB_PORT", " 445 ")

	cfg, _ := LoadFromEnv()

	// strconv.Atoi will fail on spaces, so it should default to 445
	if cfg.Port != 445 {
		t.Errorf("Expected default port 445 for invalid port string, got %d", cfg.Port)
	}
}

// Test case sensitivity of auth protocol
func TestLoadFromEnv_AuthProtocolCaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"uppercase NEGOTIATE", "NEGOTIATE", "negotiate"},
		{"mixed case Negotiate", "Negotiate", "negotiate"},
		{"uppercase NTLM", "NTLM", "ntlm"},
		{"mixed case Ntlm", "Ntlm", "ntlm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("SMB_SERVER_NAME", "test")
			os.Setenv("SMB_SERVER_IP", "127.0.0.1")
			os.Setenv("SMB_SHARE_NAME", "share")
			os.Setenv("SMB_USERNAME", "user")
			os.Setenv("SMB_PASSWORD", "pass")
			os.Setenv("SMB_AUTH_PROTOCOL", tt.value)

			cfg, _ := LoadFromEnv()

			if cfg.AuthProtocol != tt.expected {
				t.Errorf("For value '%s', expected AuthProtocol='%s', got '%s'", tt.value, tt.expected, cfg.AuthProtocol)
			}
		})
	}
}

// Test that domain is properly set
func TestLoadFromEnv_DomainPersistence(t *testing.T) {
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "test")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "share")
	os.Setenv("SMB_USERNAME", "user")
	os.Setenv("SMB_PASSWORD", "pass")
	os.Setenv("SMB_DOMAIN", "MYDOMAIN")

	cfg, _ := LoadFromEnv()

	if cfg.Domain != "MYDOMAIN" {
		t.Errorf("Expected Domain='MYDOMAIN', got '%s'", cfg.Domain)
	}
}

// Test empty domain
func TestLoadFromEnv_EmptyDomain(t *testing.T) {
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "test")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "share")
	os.Setenv("SMB_USERNAME", "user")
	os.Setenv("SMB_PASSWORD", "pass")

	cfg, _ := LoadFromEnv()

	if cfg.Domain != "" {
		t.Errorf("Expected empty Domain, got '%s'", cfg.Domain)
	}
}
