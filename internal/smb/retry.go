package smb

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bancey/document-smbrelay-service/internal/config"
	"github.com/bancey/document-smbrelay-service/internal/logger"
)

// isRetryableError determines if an error is transient and should be retried
func isRetryableError(err error, output string) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	outputStr := strings.ToLower(output)

	// Network-related errors that are typically transient
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"timeout",
		"i/o timeout",
		"network is unreachable",
		"no route to host",
		"broken pipe",
		"nt_status_io_timeout",
		"nt_status_connection_refused",
		"nt_status_network_unreachable",
		"nt_status_host_unreachable",
		"nt_status_connection_reset",
		"nt_status_pipe_broken",
		"nt_status_pipe_disconnected",
		"temporary failure",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorStr, pattern) || strings.Contains(outputStr, pattern) {
			return true
		}
	}

	// Non-retryable errors (authentication, permission, file-not-found, etc.)
	nonRetryablePatterns := []string{
		"nt_status_logon_failure",
		"nt_status_access_denied",
		"nt_status_bad_network_name",
		"nt_status_object_name_not_found",
		"nt_status_object_path_not_found",
		"nt_status_object_name_collision",
		"nt_status_file_is_a_directory",
		"authentication failed",
		"invalid credentials",
		"access denied",
		"not found",
		"invalid parameter",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errorStr, pattern) || strings.Contains(outputStr, pattern) {
			return false
		}
	}

	// Default: don't retry unknown errors
	return false
}

// calculateBackoff calculates the delay for the next retry using exponential backoff
func calculateBackoff(attempt int, cfg *config.SMBConfig) time.Duration {
	// Calculate exponential backoff: initialDelay * (backoff ^ attempt)
	delay := cfg.InitialRetryDelay * math.Pow(cfg.RetryBackoff, float64(attempt))

	// Cap at maximum delay
	if delay > cfg.MaxRetryDelay {
		delay = cfg.MaxRetryDelay
	}

	return time.Duration(delay * float64(time.Second))
}

// executeWithRetry executes a function with retry logic for transient errors
func executeWithRetry(
	operation string,
	cfg *config.SMBConfig,
	fn func() (string, error),
) (string, error) {
	var lastOutput string
	var lastErr error

	maxAttempts := cfg.MaxRetries + 1 // +1 for initial attempt

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Execute the operation
		output, err := fn()

		// If successful, return immediately
		if err == nil {
			if attempt > 0 {
				logger.Info(fmt.Sprintf("%s succeeded after %d retries", operation, attempt))
			}
			return output, nil
		}

		// Store the error for potential retry
		lastOutput = output
		lastErr = err

		// Check if this is the last attempt
		if attempt == maxAttempts-1 {
			// No more retries available
			if attempt > 0 {
				logger.Error(fmt.Sprintf("%s failed after %d retries: %v", operation, attempt, err))
			}
			break
		}

		// Check if the error is retryable
		if !isRetryableError(err, output) {
			// Non-retryable error, fail immediately
			if attempt > 0 {
				logger.Info(fmt.Sprintf("%s failed with non-retryable error after %d attempts: %v", operation, attempt+1, err))
			}
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, cfg)

		// Log retry attempt
		logger.Info(fmt.Sprintf("%s failed (attempt %d/%d), retrying in %v: %v",
			operation, attempt+1, maxAttempts, delay, err))

		// Wait before retrying
		time.Sleep(delay)
	}

	return lastOutput, lastErr
}
