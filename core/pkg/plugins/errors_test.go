package plugins

import (
	"testing"

	"ntels.com/pharos/core/external"
)

// TestIsDatasourceUnavailable tests the IsDatasourceUnavailable function
func TestIsDatasourceUnavailable(t *testing.T) {
	// Test with empty string
	t.Run("EmptyString", func(t *testing.T) {
		result := IsDatasourceUnavailable("")
		if result {
			t.Error("Expected false for empty string, got true")
		}
	})

	// Test with ErrorNotExistDatasource
	t.Run("ErrorNotExistDatasource", func(t *testing.T) {
		err := external.ErrorNotExistDatasource.Error()
		result := IsDatasourceUnavailable(err)
		if !result {
			t.Errorf("Expected true for ErrorNotExistDatasource (%s), got false", err)
		}
	})

	// Test with ErrorNotSupportedDatasourceType
	t.Run("ErrorNotSupportedDatasourceType", func(t *testing.T) {
		err := external.ErrorNotSupportedDatasourceType.Error()
		result := IsDatasourceUnavailable(err)
		if !result {
			t.Errorf("Expected true for ErrorNotSupportedDatasourceType (%s), got false", err)
		}
	})

	// Test with ErrorNotSupportedDataServer
	t.Run("ErrorNotSupportedDataServer", func(t *testing.T) {
		err := external.ErrorNotSupportedDataServer.Error()
		result := IsDatasourceUnavailable(err)
		if !result {
			t.Errorf("Expected true for ErrorNotSupportedDataServer (%s), got false", err)
		}
	})

	// Test with ErrorNotSupported
	t.Run("ErrorNotSupported", func(t *testing.T) {
		err := external.ErrorNotSupported.Error()
		result := IsDatasourceUnavailable(err)
		if !result {
			t.Errorf("Expected true for ErrorNotSupported (%s), got false", err)
		}
	})

	// Test with ErrorNotImplemented
	t.Run("ErrorNotImplemented", func(t *testing.T) {
		err := external.ErrorNotImplemented.Error()
		result := IsDatasourceUnavailable(err)
		if !result {
			t.Errorf("Expected true for ErrorNotImplemented (%s), got false", err)
		}
	})

	// Test with other error messages
	t.Run("OtherErrors", func(t *testing.T) {
		otherErrors := []string{
			"connection timeout",
			"invalid credentials",
			"database not found",
			"syntax error",
			"permission denied",
			"network error",
		}

		for _, errMsg := range otherErrors {
			result := IsDatasourceUnavailable(errMsg)
			if result {
				t.Errorf("Expected false for other error (%s), got true", errMsg)
			}
		}
	})

	// Test with partial matches (should return false)
	t.Run("PartialMatches", func(t *testing.T) {
		partialMatches := []string{
			"datasource not exist but not exact",
			"not supported datasource type with extra text",
			"prefix " + external.ErrorNotSupported.Error(),
			external.ErrorNotImplemented.Error() + " suffix",
		}

		for _, errMsg := range partialMatches {
			result := IsDatasourceUnavailable(errMsg)
			if result {
				t.Errorf("Expected false for partial match (%s), got true", errMsg)
			}
		}
	})

	// Test case sensitivity
	t.Run("CaseSensitivity", func(t *testing.T) {
		upperCaseError := "DATASOURCE NOT EXIST"
		result := IsDatasourceUnavailable(upperCaseError)
		if result {
			t.Errorf("Expected false for case-different error (%s), got true", upperCaseError)
		}
	})

	// Test with whitespace variations
	t.Run("WhitespaceVariations", func(t *testing.T) {
		whitespaceErrors := []string{
			" " + external.ErrorNotExistDatasource.Error(),
			external.ErrorNotSupported.Error() + " ",
			"\t" + external.ErrorNotImplemented.Error() + "\n",
		}

		for _, errMsg := range whitespaceErrors {
			result := IsDatasourceUnavailable(errMsg)
			if result {
				t.Errorf("Expected false for whitespace variation (%s), got true", errMsg)
			}
		}
	})
}

// TestIsDatasourceUnavailable_AllKnownErrors tests all known unavailable errors
func TestIsDatasourceUnavailable_AllKnownErrors(t *testing.T) {
	knownUnavailableErrors := []error{
		external.ErrorNotExistDatasource,
		external.ErrorNotSupportedDatasourceType,
		external.ErrorNotSupportedDataServer,
		external.ErrorNotSupported,
		external.ErrorNotImplemented,
	}

	for _, err := range knownUnavailableErrors {
		errorString := err.Error()
		result := IsDatasourceUnavailable(errorString)
		if !result {
			t.Errorf("Expected true for known unavailable error (%s), got false", errorString)
		}
	}
}

// TestIsDatasourceUnavailable_EdgeCases tests edge cases
func TestIsDatasourceUnavailable_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"NilEquivalent", "", false},
		{"SingleSpace", " ", false},
		{"OnlyNewlines", "\n\n", false},
		{"OnlyTabs", "\t\t", false},
		{"MixedWhitespace", " \t\n ", false},
		{"ExactErrorMatch", external.ErrorNotExistDatasource.Error(), true},
		{"UnicodeCharacters", "データソースが存在しません", false},
		{"SpecialCharacters", "!@#$%^&*()", false},
		{"NumericString", "12345", false},
		{"BooleanString", "true", false},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDatasourceUnavailable(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v for input (%s), got %v", tc.expected, tc.input, result)
			}
		})
	}
}

// TestIsDatasourceUnavailable_Performance tests performance with repeated calls
func TestIsDatasourceUnavailable_Performance(t *testing.T) {
	testErrors := []string{
		external.ErrorNotExistDatasource.Error(),
		external.ErrorNotSupportedDatasourceType.Error(),
		"some other error",
		"",
		external.ErrorNotImplemented.Error(),
	}

	// This test ensures the function doesn't have performance issues
	for range 1000 {
		for _, err := range testErrors {
			_ = IsDatasourceUnavailable(err)
		}
	}

	t.Log("Performance test completed successfully")
}

// TestIsDatasourceUnavailable_Consistency tests that function returns consistent results
func TestIsDatasourceUnavailable_Consistency(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{external.ErrorNotExistDatasource.Error(), true},
		{external.ErrorNotSupportedDatasourceType.Error(), true},
		{external.ErrorNotSupportedDataServer.Error(), true},
		{external.ErrorNotSupported.Error(), true},
		{external.ErrorNotImplemented.Error(), true},
		{"other error", false},
		{"", false},
	}

	// Call each test case multiple times to ensure consistency
	for _, tc := range testCases {
		for i := range 10 {
			result := IsDatasourceUnavailable(tc.input)
			if result != tc.expected {
				t.Errorf("Inconsistent result for input (%s) on iteration %d: expected %v, got %v",
					tc.input, i, tc.expected, result)
			}
		}
	}
}

// TestIsDatasourceUnavailable_LongStrings tests with very long error strings
func TestIsDatasourceUnavailable_LongStrings(t *testing.T) {
	// Test with a very long string that doesn't match
	longString := "This is a very long error message that should not match any of the known unavailable error types. " +
		"It contains many words and characters but none of them should trigger the unavailable datasource detection. " +
		"We want to make sure that the function can handle long strings efficiently without any issues."

	result := IsDatasourceUnavailable(longString)
	if result {
		t.Error("Expected false for long non-matching string, got true")
	}

	// Test with a long string that contains a matching error at the end
	longStringWithMatch := longString + " " + external.ErrorNotExistDatasource.Error()
	result = IsDatasourceUnavailable(longStringWithMatch)
	if result {
		t.Error("Expected false for long string with non-exact match, got true")
	}

	// Test with exact match that happens to be long (if any of the errors were long)
	exactMatch := external.ErrorNotSupportedDatasourceType.Error()
	result = IsDatasourceUnavailable(exactMatch)
	if !result {
		t.Error("Expected true for exact match of ErrorNotSupportedDatasourceType, got false")
	}
}

// Benchmark tests
func BenchmarkIsDatasourceUnavailable_TrueCase(b *testing.B) {
	errMsg := external.ErrorNotExistDatasource.Error()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsDatasourceUnavailable(errMsg)
	}
}

func BenchmarkIsDatasourceUnavailable_FalseCase(b *testing.B) {
	errMsg := "some random error message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsDatasourceUnavailable(errMsg)
	}
}

func BenchmarkIsDatasourceUnavailable_EmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsDatasourceUnavailable("")
	}
}

func BenchmarkIsDatasourceUnavailable_AllErrorTypes(b *testing.B) {
	errors := []string{
		external.ErrorNotExistDatasource.Error(),
		external.ErrorNotSupportedDatasourceType.Error(),
		external.ErrorNotSupportedDataServer.Error(),
		external.ErrorNotSupported.Error(),
		external.ErrorNotImplemented.Error(),
		"other error",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, err := range errors {
			_ = IsDatasourceUnavailable(err)
		}
	}
}
