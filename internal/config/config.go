// Package config handles SMB configuration management
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SMBConfig holds the SMB server configuration
// Fields are ordered for optimal memory alignment
type SMBConfig struct {
	ServerName   string
	ServerIP     string
	ShareName    string
	Username     string
	Password     string
	Domain       string
	AuthProtocol string
	Port         int
	UseNTLMv2    bool
}

// LoadFromEnv loads SMB configuration from environment variables
// Returns the config and a list of missing required variables
func LoadFromEnv() (*SMBConfig, []string) {
	serverName := os.Getenv("SMB_SERVER_NAME")
	serverIP := os.Getenv("SMB_SERVER_IP")
	shareName := os.Getenv("SMB_SHARE_NAME")
	username := os.Getenv("SMB_USERNAME")
	password := os.Getenv("SMB_PASSWORD")
	domain := os.Getenv("SMB_DOMAIN")

	portStr := os.Getenv("SMB_PORT")
	if portStr == "" {
		portStr = "445"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 445
	}

	useNTLMv2Str := strings.ToLower(os.Getenv("SMB_USE_NTLM_V2"))
	if useNTLMv2Str == "" {
		useNTLMv2Str = "true"
	}
	useNTLMv2 := useNTLMv2Str == "true" || useNTLMv2Str == "1" || useNTLMv2Str == "yes"

	// Auth protocol: negotiate, ntlm, or kerberos
	authProtocol := strings.ToLower(os.Getenv("SMB_AUTH_PROTOCOL"))
	validProtocols := map[string]bool{
		"negotiate": true,
		"ntlm":      true,
		"kerberos":  true,
	}

	if authProtocol == "" || !validProtocols[authProtocol] {
		// If not explicitly set, derive from useNTLMv2 for backward compatibility
		if useNTLMv2 {
			authProtocol = "ntlm"
		} else {
			authProtocol = "negotiate"
		}
	}

	config := &SMBConfig{
		ServerName:   serverName,
		ServerIP:     serverIP,
		ShareName:    shareName,
		Username:     username,
		Password:     password,
		Domain:       domain,
		Port:         port,
		UseNTLMv2:    useNTLMv2,
		AuthProtocol: authProtocol,
	}

	// Check required fields
	var missing []string
	if serverName == "" {
		missing = append(missing, "SMB_SERVER_NAME")
	}
	if serverIP == "" {
		missing = append(missing, "SMB_SERVER_IP")
	}
	if shareName == "" {
		missing = append(missing, "SMB_SHARE_NAME")
	}

	// Username and password are only required for non-Kerberos authentication
	if authProtocol != "kerberos" {
		if username == "" {
			missing = append(missing, "SMB_USERNAME")
		}
		if password == "" {
			missing = append(missing, "SMB_PASSWORD")
		}
	}

	return config, missing
}

// GetServer returns the server address with port
func (c *SMBConfig) GetServer() string {
	server := c.ServerIP
	if server == "" {
		server = c.ServerName
	}
	return fmt.Sprintf("%s:%d", server, c.Port)
}

// GetServerDisplay returns a display string for the server
func (c *SMBConfig) GetServerDisplay() string {
	return fmt.Sprintf("%s (%s:%d)", c.ServerName, c.ServerIP, c.Port)
}
