package testutil

import (
	"testing"
)

// TestAssertContains_Success tests AssertContains with matching string
func TestAssertContains_Success(t *testing.T) {
	// Create a mock testing.T to capture failures
	mockT := &testing.T{}

	AssertContains(mockT, "hello world", "world")

	// If no error, the assertion passed
	if mockT.Failed() {
		t.Error("AssertContains failed when it should have passed")
	}
}

// TestAssertContains_Failure tests AssertContains with non-matching string
func TestAssertContains_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertContains(mockT, "hello world", "xyz")

	// Should have failed because substring not found
	if !mockT.Failed() {
		t.Error("AssertContains should have failed for non-matching substring")
	}
}

// TestAssertNotContains_Success tests AssertNotContains with non-matching string
func TestAssertNotContains_Success(t *testing.T) {
	mockT := &testing.T{}

	AssertNotContains(mockT, "hello world", "xyz")

	if mockT.Failed() {
		t.Error("AssertNotContains failed when it should have passed")
	}
}

// TestAssertNotContains_Failure tests AssertNotContains with matching string
func TestAssertNotContains_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertNotContains(mockT, "hello world", "world")

	// Should have failed because substring was found
	if !mockT.Failed() {
		t.Error("AssertNotContains should have failed for matching substring")
	}
}

// TestAssertEqual_Success tests AssertEqual with equal values
func TestAssertEqual_Success(t *testing.T) {
	mockT := &testing.T{}

	AssertEqual(mockT, 42, 42)

	if mockT.Failed() {
		t.Error("AssertEqual failed when it should have passed")
	}
}

// TestAssertEqual_Failure tests AssertEqual with unequal values
func TestAssertEqual_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertEqual(mockT, 42, 43)

	// Should have failed because values are different
	if !mockT.Failed() {
		t.Error("AssertEqual should have failed for unequal values")
	}
}

// TestAssertNotEqual_Success tests AssertNotEqual with different values
func TestAssertNotEqual_Success(t *testing.T) {
	mockT := &testing.T{}

	AssertNotEqual(mockT, 42, 43)

	if mockT.Failed() {
		t.Error("AssertNotEqual failed when it should have passed")
	}
}

// TestAssertNotEqual_Failure tests AssertNotEqual with equal values
func TestAssertNotEqual_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertNotEqual(mockT, 42, 42)

	// Should have failed because values are equal
	if !mockT.Failed() {
		t.Error("AssertNotEqual should have failed for equal values")
	}
}

// TestAssertTrue_Success tests AssertTrue with true condition
func TestAssertTrue_Success(t *testing.T) {
	mockT := &testing.T{}

	AssertTrue(mockT, true)

	if mockT.Failed() {
		t.Error("AssertTrue failed when it should have passed")
	}
}

// TestAssertTrue_Failure tests AssertTrue with false condition
func TestAssertTrue_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertTrue(mockT, false)

	// Should have failed because condition is false
	if !mockT.Failed() {
		t.Error("AssertTrue should have failed for false condition")
	}
}

// TestAssertFalse_Success tests AssertFalse with false condition
func TestAssertFalse_Success(t *testing.T) {
	mockT := &testing.T{}

	AssertFalse(mockT, false)

	if mockT.Failed() {
		t.Error("AssertFalse failed when it should have passed")
	}
}

// TestAssertFalse_Failure tests AssertFalse with true condition
func TestAssertFalse_Failure(t *testing.T) {
	mockT := &testing.T{}

	AssertFalse(mockT, true)

	// Should have failed because condition is true
	if !mockT.Failed() {
		t.Error("AssertFalse should have failed for true condition")
	}
}

// TestAssertNil_Success tests AssertNil with nil value
func TestAssertNil_Success(t *testing.T) {
	mockT := &testing.T{}

	var nilValue error = nil
	AssertNil(mockT, nilValue)

	if mockT.Failed() {
		t.Error("AssertNil failed when it should have passed")
	}
}

// TestAssertNil_Failure tests AssertNil with non-nil value
func TestAssertNil_Failure(t *testing.T) {
	mockT := &testing.T{}

	type testError struct{}
	nonNilValue := &testError{}
	AssertNil(mockT, nonNilValue)

	// Should have failed because value is not nil
	if !mockT.Failed() {
		t.Error("AssertNil should have failed for non-nil value")
	}
}

// TestAssertNotNil_Success tests AssertNotNil with non-nil value
func TestAssertNotNil_Success(t *testing.T) {
	mockT := &testing.T{}

	nonNilValue := "not nil"
	AssertNotNil(mockT, nonNilValue)

	if mockT.Failed() {
		t.Error("AssertNotNil failed when it should have passed")
	}
}

// TestAssertNotNil_Failure tests AssertNotNil with nil value
func TestAssertNotNil_Failure(t *testing.T) {
	mockT := &testing.T{}

	var nilValue interface{}
	AssertNotNil(mockT, nilValue)

	// Should have failed because value is nil
	if !mockT.Failed() {
		t.Error("AssertNotNil should have failed for nil value")
	}
}

// TestContainsInner_EmptyNeedle tests containsInner with empty needle
func TestContainsInner_EmptyNeedle(t *testing.T) {
	result := containsInner("hello world", "")

	// Empty needle should be found (Go strings.Contains behavior)
	if !result {
		t.Error("Expected empty needle to be found")
	}
}

// TestContainsInner_EmptyHaystack tests containsInner with empty haystack
func TestContainsInner_EmptyHaystack(t *testing.T) {
	result := containsInner("", "hello")

	// Non-empty needle in empty haystack should not be found
	if result {
		t.Error("Expected needle not to be found in empty haystack")
	}
}

// TestContainsInner_BothEmpty tests containsInner with both empty strings
func TestContainsInner_BothEmpty(t *testing.T) {
	result := containsInner("", "")

	// Empty needle in empty haystack should be found
	if !result {
		t.Error("Expected empty needle to be found in empty haystack")
	}
}

// TestContains_DirectCall tests the contains function directly
func TestContains_DirectCall(t *testing.T) {
	tests := []struct {
		name     string
		haystack string
		needle   string
		expected bool
	}{
		{"substring in middle", "hello world", "world", true},
		{"substring not found", "hello world", "xyz", false},
		{"exact match", "hello", "hello", true},
		{"empty needle", "hello", "", true},
		{"empty haystack", "", "hello", false},
		{"both empty", "", "", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.haystack, tt.needle)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, expected %v",
					tt.haystack, tt.needle, result, tt.expected)
			}
		})
	}
}
