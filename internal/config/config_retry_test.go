package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv_RetryConfiguration(t *testing.T) {
	tests := []struct {
		envVars              map[string]string
		name                 string
		expectedInitialDelay float64
		expectedMaxDelay     float64
		expectedBackoff      float64
		expectedMaxRetries   int
	}{
		{
			name: "Default retry values",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "testserver",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_SHARE_NAME":  "testshare",
				"SMB_USERNAME":    "testuser",
				"SMB_PASSWORD":    "testpass",
			},
			expectedMaxRetries:   defaultMaxRetries,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      defaultRetryBackoff,
		},
		{
			name: "Custom max retries",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "testserver",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_SHARE_NAME":  "testshare",
				"SMB_USERNAME":    "testuser",
				"SMB_PASSWORD":    "testpass",
				"SMB_MAX_RETRIES": "5",
			},
			expectedMaxRetries:   5,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      defaultRetryBackoff,
		},
		{
			name: "Custom retry delays",
			envVars: map[string]string{
				"SMB_SERVER_NAME":         "testserver",
				"SMB_SERVER_IP":           "127.0.0.1",
				"SMB_SHARE_NAME":          "testshare",
				"SMB_USERNAME":            "testuser",
				"SMB_PASSWORD":            "testpass",
				"SMB_RETRY_INITIAL_DELAY": "0.5",
				"SMB_RETRY_MAX_DELAY":     "60.0",
			},
			expectedMaxRetries:   defaultMaxRetries,
			expectedInitialDelay: 0.5,
			expectedMaxDelay:     60.0,
			expectedBackoff:      defaultRetryBackoff,
		},
		{
			name: "Custom backoff multiplier",
			envVars: map[string]string{
				"SMB_SERVER_NAME":   "testserver",
				"SMB_SERVER_IP":     "127.0.0.1",
				"SMB_SHARE_NAME":    "testshare",
				"SMB_USERNAME":      "testuser",
				"SMB_PASSWORD":      "testpass",
				"SMB_RETRY_BACKOFF": "1.5",
			},
			expectedMaxRetries:   defaultMaxRetries,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      1.5,
		},
		{
			name: "All retry settings custom",
			envVars: map[string]string{
				"SMB_SERVER_NAME":         "testserver",
				"SMB_SERVER_IP":           "127.0.0.1",
				"SMB_SHARE_NAME":          "testshare",
				"SMB_USERNAME":            "testuser",
				"SMB_PASSWORD":            "testpass",
				"SMB_MAX_RETRIES":         "10",
				"SMB_RETRY_INITIAL_DELAY": "2.0",
				"SMB_RETRY_MAX_DELAY":     "120.0",
				"SMB_RETRY_BACKOFF":       "3.0",
			},
			expectedMaxRetries:   10,
			expectedInitialDelay: 2.0,
			expectedMaxDelay:     120.0,
			expectedBackoff:      3.0,
		},
		{
			name: "Invalid retry values default to defaults",
			envVars: map[string]string{
				"SMB_SERVER_NAME":         "testserver",
				"SMB_SERVER_IP":           "127.0.0.1",
				"SMB_SHARE_NAME":          "testshare",
				"SMB_USERNAME":            "testuser",
				"SMB_PASSWORD":            "testpass",
				"SMB_MAX_RETRIES":         "invalid",
				"SMB_RETRY_INITIAL_DELAY": "invalid",
				"SMB_RETRY_MAX_DELAY":     "invalid",
				"SMB_RETRY_BACKOFF":       "invalid",
			},
			expectedMaxRetries:   defaultMaxRetries,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      defaultRetryBackoff,
		},
		{
			name: "Negative retry values default to defaults",
			envVars: map[string]string{
				"SMB_SERVER_NAME":         "testserver",
				"SMB_SERVER_IP":           "127.0.0.1",
				"SMB_SHARE_NAME":          "testshare",
				"SMB_USERNAME":            "testuser",
				"SMB_PASSWORD":            "testpass",
				"SMB_MAX_RETRIES":         "-5",
				"SMB_RETRY_INITIAL_DELAY": "-1.0",
				"SMB_RETRY_MAX_DELAY":     "-10.0",
				"SMB_RETRY_BACKOFF":       "-2.0",
			},
			expectedMaxRetries:   defaultMaxRetries,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      defaultRetryBackoff,
		},
		{
			name: "Zero retries is valid",
			envVars: map[string]string{
				"SMB_SERVER_NAME": "testserver",
				"SMB_SERVER_IP":   "127.0.0.1",
				"SMB_SHARE_NAME":  "testshare",
				"SMB_USERNAME":    "testuser",
				"SMB_PASSWORD":    "testpass",
				"SMB_MAX_RETRIES": "0",
			},
			expectedMaxRetries:   0,
			expectedInitialDelay: defaultInitialRetryDelay,
			expectedMaxDelay:     defaultMaxRetryDelay,
			expectedBackoff:      defaultRetryBackoff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			// Load config
			cfg, missing := LoadFromEnv()

			// Check no missing required fields
			if len(missing) > 0 {
				t.Errorf("Unexpected missing fields: %v", missing)
			}

			// Check retry configuration
			if cfg.MaxRetries != tt.expectedMaxRetries {
				t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, tt.expectedMaxRetries)
			}
			if cfg.InitialRetryDelay != tt.expectedInitialDelay {
				t.Errorf("InitialRetryDelay = %f, want %f", cfg.InitialRetryDelay, tt.expectedInitialDelay)
			}
			if cfg.MaxRetryDelay != tt.expectedMaxDelay {
				t.Errorf("MaxRetryDelay = %f, want %f", cfg.MaxRetryDelay, tt.expectedMaxDelay)
			}
			if cfg.RetryBackoff != tt.expectedBackoff {
				t.Errorf("RetryBackoff = %f, want %f", cfg.RetryBackoff, tt.expectedBackoff)
			}
		})
	}
}
