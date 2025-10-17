// Package config handles SMB configuration management
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPortStr        = "445"
	defaultPort           = 445
	trueValue             = "true"
	oneValue              = "1"
	yesValue              = "yes"
	authProtocolNTLM      = "ntlm"
	authProtocolNegotiate = "negotiate"
	authProtocolKerberos  = "kerberos"
)

// SMBConfig holds the SMB server configuration
// Fields are ordered for optimal memory alignment
type SMBConfig struct {
	ServerName     string
	ServerIP       string
	ShareName      string
	BasePath       string // Base path within the share (e.g., "apps/myapp")
	Username       string
	Password       string
	Domain         string
	AuthProtocol   string
	Port           int
	UseNTLMv2      bool
	LogSmbCommands bool
}

// parseBoolEnv parses a boolean environment variable
func parseBoolEnv(value string) bool {
	value = strings.ToLower(value)
	return value == trueValue || value == oneValue || value == yesValue
}

// getPortFromEnv gets the port from environment variable with fallback
func getPortFromEnv() int {
	portStr := os.Getenv("SMB_PORT")
	if portStr == "" {
		portStr = defaultPortStr
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = defaultPort
	}
	return port
}

// getAuthProtocol determines the authentication protocol
func getAuthProtocol(useNTLMv2 bool) string {
	authProtocol := strings.ToLower(os.Getenv("SMB_AUTH_PROTOCOL"))
	validProtocols := map[string]bool{
		authProtocolNegotiate: true,
		authProtocolNTLM:      true,
		authProtocolKerberos:  true,
	}

	if authProtocol == "" || !validProtocols[authProtocol] {
		// If not explicitly set, derive from useNTLMv2 for backward compatibility
		if useNTLMv2 {
			authProtocol = authProtocolNTLM
		} else {
			authProtocol = authProtocolNegotiate
		}
	}
	return authProtocol
}

// LoadFromEnv loads SMB configuration from environment variables
// Returns the config and a list of missing required variables
func LoadFromEnv() (*SMBConfig, []string) {
	serverName := os.Getenv("SMB_SERVER_NAME")
	serverIP := os.Getenv("SMB_SERVER_IP")
	shareName := os.Getenv("SMB_SHARE_NAME")
	basePath := os.Getenv("SMB_BASE_PATH")
	username := os.Getenv("SMB_USERNAME")
	password := os.Getenv("SMB_PASSWORD")
	domain := os.Getenv("SMB_DOMAIN")

	port := getPortFromEnv()

	useNTLMv2Str := strings.ToLower(os.Getenv("SMB_USE_NTLM_V2"))
	if useNTLMv2Str == "" {
		useNTLMv2Str = trueValue
	}
	useNTLMv2 := parseBoolEnv(useNTLMv2Str)

	// Log SMB commands for debugging (support both env var names for user convenience)
	logSmbCommandsStr := os.Getenv("LOG_SMB_COMMANDS")
	if logSmbCommandsStr == "" {
		logSmbCommandsStr = os.Getenv("SMB_LOG_COMMANDS") // Alternative name
	}
	logSmbCommands := parseBoolEnv(logSmbCommandsStr)

	// Auth protocol: negotiate, ntlm, or kerberos
	authProtocol := getAuthProtocol(useNTLMv2)

	config := &SMBConfig{
		ServerName:     serverName,
		ServerIP:       serverIP,
		ShareName:      shareName,
		BasePath:       basePath,
		Username:       username,
		Password:       password,
		Domain:         domain,
		Port:           port,
		UseNTLMv2:      useNTLMv2,
		AuthProtocol:   authProtocol,
		LogSmbCommands: logSmbCommands,
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
	if authProtocol != authProtocolKerberos {
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
