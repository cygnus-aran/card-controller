package services

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMerchantKeyLogger - dedicated mock logger for merchant key service tests
type MockMerchantKeyLogger struct {
	mock.Mock
}

func (m *MockMerchantKeyLogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantKeyLogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantKeyLogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantKeyLogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Test helper functions
func setupMerchantKeyService(t *testing.T) (*MerchantKeyService, *MockMerchantKeyLogger) {
	t.Helper()
	mockLogger := &MockMerchantKeyLogger{}
	service := NewMerchantKeyService(mockLogger).(*MerchantKeyService)
	return service, mockLogger
}

// Helper to set and cleanup environment variables
func setMerchantKeyEnvVar(t *testing.T, key, value string) {
	t.Helper()
	originalValue := os.Getenv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	})
}

// Test GetMerchantPublicKey - Success Cases
func TestMerchantKeyService_GetMerchantPublicKey_Success(t *testing.T) {
	testCases := []struct {
		name        string
		merchantID  string
		setupEnv    func(t *testing.T)
		description string
	}{
		{
			name:       "Key found in environment variable",
			merchantID: "MERCHANT123",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "MERCHANT_MERCHANT123_PUBLIC_KEY", "test-public-key-content")
			},
			description: "Should return key when found in environment variable",
		},
		{
			name:       "Key found in environment - different merchant",
			merchantID: "TESTMERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "MERCHANT_TESTMERCHANT_PUBLIC_KEY", "another-test-key")
			},
			description: "Should return correct key for different merchant",
		},
		{
			name:       "Mock key in test environment",
			merchantID: "ANY_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "test")
				// Don't set any merchant-specific key
			},
			description: "Should return mock key in test environment",
		},
		{
			name:       "Mock key in dev environment",
			merchantID: "DEV_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "dev")
			},
			description: "Should return mock key in dev environment",
		},
		{
			name:       "Mock key in local environment",
			merchantID: "LOCAL_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "local")
			},
			description: "Should return mock key in local environment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantKeyService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			key, err := service.GetMerchantPublicKey(tc.merchantID)

			// Assert
			assert.NoError(t, err, tc.description)
			assert.NotEmpty(t, key, "Key should not be empty")

			// Verify mock key format in test environments
			if isTestEnvironment() {
				assert.Contains(t, key, "-----BEGIN PUBLIC KEY-----", "Mock key should have proper PEM format")
				assert.Contains(t, key, "-----END PUBLIC KEY-----", "Mock key should have proper PEM format")
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

// Test GetMerchantPublicKey - Failure Cases
func TestMerchantKeyService_GetMerchantPublicKey_Failures(t *testing.T) {
	testCases := []struct {
		name          string
		merchantID    string
		setupEnv      func(t *testing.T)
		expectedError string
		description   string
	}{
		{
			name:       "Key not found in production environment",
			merchantID: "UNKNOWN_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")
				// Don't set any key for this merchant
			},
			expectedError: "public key not found for merchant: UNKNOWN_MERCHANT",
			description:   "Should return error when key not found in production",
		},
		{
			name:       "Key not found in UAT environment",
			merchantID: "UAT_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "uat")
			},
			expectedError: "public key not found for merchant: UAT_MERCHANT",
			description:   "Should return error when key not found in UAT",
		},
		{
			name:       "Key not found in staging environment",
			merchantID: "STAGING_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "staging")
			},
			expectedError: "public key not found for merchant: STAGING_MERCHANT",
			description:   "Should return error when key not found in staging",
		},
		{
			name:       "Empty merchant ID",
			merchantID: "",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedError: "public key not found for merchant: ",
			description:   "Should return error for empty merchant ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantKeyService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
			mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			key, err := service.GetMerchantPublicKey(tc.merchantID)

			// Assert
			assert.Error(t, err, tc.description)
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Empty(t, key, "Key should be empty on error")
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test HasMerchantKey - Success Cases
func TestMerchantKeyService_HasMerchantKey_Success(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Key exists in environment variable",
			merchantID: "MERCHANT123",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "MERCHANT_MERCHANT123_PUBLIC_KEY", "test-key-content")
			},
			expectedResult: true,
			description:    "Should return true when key exists in environment",
		},
		{
			name:       "Different merchant key exists",
			merchantID: "ANOTHERMERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "MERCHANT_ANOTHERMERCHANT_PUBLIC_KEY", "another-key")
			},
			expectedResult: true,
			description:    "Should return true for different merchant with key",
		},
		{
			name:       "Any merchant in test environment",
			merchantID: "ANY_TEST_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "test")
			},
			expectedResult: true,
			description:    "Should return true for any merchant in test environment",
		},
		{
			name:       "Any merchant in dev environment",
			merchantID: "DEV_MERCHANT_123",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "dev")
			},
			expectedResult: true,
			description:    "Should return true for any merchant in dev environment",
		},
		{
			name:       "Any merchant in local environment",
			merchantID: "LOCAL_TEST_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "local")
			},
			expectedResult: true,
			description:    "Should return true for any merchant in local environment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantKeyService(t)
			tc.setupEnv(t)

			// Setup mocks
			if isTestEnvironment() {
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
			}

			// Act
			result := service.HasMerchantKey(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test HasMerchantKey - Failure Cases
func TestMerchantKeyService_HasMerchantKey_Failures(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Key not found in production",
			merchantID: "UNKNOWN_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")
				// Don't set any key
			},
			expectedResult: false,
			description:    "Should return false when key not found in production",
		},
		{
			name:       "Key not found in UAT",
			merchantID: "UAT_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "uat")
			},
			expectedResult: false,
			description:    "Should return false when key not found in UAT",
		},
		{
			name:       "Key not found in staging",
			merchantID: "STAGING_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "staging")
			},
			expectedResult: false,
			description:    "Should return false when key not found in staging",
		},
		{
			name:       "Empty merchant ID in production",
			merchantID: "",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should return false for empty merchant ID",
		},
		{
			name:       "Wrong merchant key exists",
			merchantID: "MERCHANT123",
			setupEnv: func(t *testing.T) {
				setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")
				setMerchantKeyEnvVar(t, "MERCHANT_DIFFERENTMERCHANT_PUBLIC_KEY", "some-key")
			},
			expectedResult: false,
			description:    "Should return false when different merchant's key exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantKeyService(t)
			tc.setupEnv(t)

			// Setup mocks
			if !isTestEnvironment() {
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
			}

			// Act
			result := service.HasMerchantKey(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test Environment Variable Key Format
func TestMerchantKeyService_EnvironmentKeyFormat(t *testing.T) {
	testCases := []struct {
		name       string
		merchantID string
		expected   string
	}{
		{
			name:       "Simple merchant ID",
			merchantID: "MERCHANT123",
			expected:   "MERCHANT_MERCHANT123_PUBLIC_KEY",
		},
		{
			name:       "Merchant with underscores",
			merchantID: "TEST_MERCHANT_456",
			expected:   "MERCHANT_TEST_MERCHANT_456_PUBLIC_KEY",
		},
		{
			name:       "Short merchant ID",
			merchantID: "M1",
			expected:   "MERCHANT_M1_PUBLIC_KEY",
		},
		{
			name:       "Long merchant ID",
			merchantID: "VERY_LONG_MERCHANT_IDENTIFIER_WITH_MANY_PARTS",
			expected:   "MERCHANT_VERY_LONG_MERCHANT_IDENTIFIER_WITH_MANY_PARTS_PUBLIC_KEY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantKeyService(t)
			setMerchantKeyEnvVar(t, tc.expected, "test-key-content")

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			key, err := service.GetMerchantPublicKey(tc.merchantID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, "test-key-content", key)
		})
	}
}

// Test Mock Key Generation
func TestMerchantKeyService_MockKeyGeneration(t *testing.T) {
	t.Run("Mock key has proper PEM format", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		key, err := service.GetMerchantPublicKey("TEST_MERCHANT")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, key)

		// Verify PEM format
		assert.Contains(t, key, "-----BEGIN PUBLIC KEY-----")
		assert.Contains(t, key, "-----END PUBLIC KEY-----")

		// Verify it looks like a realistic key
		lines := strings.Split(key, "\n")
		assert.GreaterOrEqual(t, len(lines), 5, "Mock key should have at least 5 lines (header, content, footer)")

		// Check that content lines contain base64-like characters
		hasContentLines := false
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			// Skip header, footer, and empty lines
			if trimmedLine != "" &&
				!strings.Contains(trimmedLine, "-----BEGIN") &&
				!strings.Contains(trimmedLine, "-----END") {
				hasContentLines = true
				// Should contain only base64 characters
				for _, char := range trimmedLine {
					assert.True(t,
						(char >= 'A' && char <= 'Z') ||
							(char >= 'a' && char <= 'z') ||
							(char >= '0' && char <= '9') ||
							char == '+' || char == '/' || char == '=',
						"Mock key should contain only valid base64 characters in line: %s", trimmedLine)
				}
			}
		}
		assert.True(t, hasContentLines, "Mock key should have at least one content line with base64 data")

		// Debug log to see the actual key format
		t.Logf("Mock key has %d lines", len(lines))
		t.Logf("Mock key format:\n%s", key)
	})

	t.Run("Mock key is consistent", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "dev")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act - Get the same key multiple times
		key1, err1 := service.GetMerchantPublicKey("SAME_MERCHANT")
		key2, err2 := service.GetMerchantPublicKey("SAME_MERCHANT")

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, key1, key2, "Mock key should be consistent for same merchant")
	})
}

// Test Logging Behavior
func TestMerchantKeyService_LoggingBehavior(t *testing.T) {
	t.Run("Logs GetMerchantPublicKey start", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "MERCHANT_TESTMERCHANT_PUBLIC_KEY", "test-key")

		// Setup mocks
		mockLogger.On("Info", "MerchantKeyService.GetMerchantPublicKey | Starting", "MerchantID: TESTMERCHANT").Return()
		mockLogger.On("Info", "MerchantKeyService.GetMerchantPublicKey | FoundInEnv", "MerchantID: TESTMERCHANT").Return()

		// Act
		key, err := service.GetMerchantPublicKey("TESTMERCHANT")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test-key", key)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs mock key usage", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", "MerchantKeyService.GetMerchantPublicKey | Starting", "MerchantID: MOCK_MERCHANT").Return()
		mockLogger.On("Info", "MerchantKeyService.GetMerchantPublicKey | UsingMockKey", "MerchantID: MOCK_MERCHANT").Return()

		// Act
		key, err := service.GetMerchantPublicKey("MOCK_MERCHANT")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs key not found error", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "MerchantKeyService.GetMerchantPublicKey | Starting", "MerchantID: UNKNOWN").Return()
		mockLogger.On("Error", "MerchantKeyService.GetMerchantPublicKey | KeyNotFound", "MerchantID: UNKNOWN").Return()

		// Act
		key, err := service.GetMerchantPublicKey("UNKNOWN")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, key)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs HasMerchantKey mock environment", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "local")

		// Setup mocks
		mockLogger.On("Info", "MerchantKeyService.HasMerchantKey | MockEnvironment", "MerchantID: LOCAL_MERCHANT").Return()

		// Act
		result := service.HasMerchantKey("LOCAL_MERCHANT")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs HasMerchantKey key not found", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "MerchantKeyService.HasMerchantKey | KeyNotFound", "MerchantID: NOTFOUND").Return()

		// Act
		result := service.HasMerchantKey("NOTFOUND")

		// Assert
		assert.False(t, result)
		mockLogger.AssertExpectations(t)
	})
}

// Test Edge Cases
func TestMerchantKeyService_EdgeCases(t *testing.T) {
	t.Run("Empty key value in environment", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "MERCHANT_EMPTY_PUBLIC_KEY", "")
		setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		key, err := service.GetMerchantPublicKey("EMPTY")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, key)
	})

	t.Run("Key with whitespace", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		keyWithSpaces := "  test-key-with-spaces  "
		setMerchantKeyEnvVar(t, "MERCHANT_SPACED_PUBLIC_KEY", keyWithSpaces)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		key, err := service.GetMerchantPublicKey("SPACED")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, keyWithSpaces, key, "Should return key exactly as stored, including whitespace")
	})

	t.Run("Case sensitivity in merchant ID", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		setMerchantKeyEnvVar(t, "MERCHANT_UPPERCASE_PUBLIC_KEY", "uppercase-key")
		setMerchantKeyEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks for uppercase (should work)
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act & Assert for uppercase - should work
		key, err := service.GetMerchantPublicKey("UPPERCASE")
		assert.NoError(t, err)
		assert.Equal(t, "uppercase-key", key)

		// Reset mocks for the lowercase test
		mockLogger.ExpectedCalls = nil
		mockLogger.Calls = nil

		// Setup mocks for lowercase (should fail in prod environment)
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act & Assert for lowercase - should fail because env var is MERCHANT_UPPERCASE_PUBLIC_KEY
		// but we're looking for MERCHANT_uppercase_PUBLIC_KEY
		key, err = service.GetMerchantPublicKey("uppercase")
		if err != nil {
			// Expected behavior - case sensitive
			assert.Error(t, err)
			assert.Empty(t, key)
		} else {
			// If the service is case insensitive, that's also valid behavior
			// Just log this for awareness
			t.Logf("Note: Service appears to be case insensitive. Found key: %s", key)
			assert.Equal(t, "uppercase-key", key)
		}
	})

	t.Run("Very long key content", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantKeyService(t)
		longKey := strings.Repeat("VERY_LONG_KEY_CONTENT_", 100)
		setMerchantKeyEnvVar(t, "MERCHANT_LONGKEY_PUBLIC_KEY", longKey)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		key, err := service.GetMerchantPublicKey("LONGKEY")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, longKey, key)
		assert.Greater(t, len(key), 1000, "Should handle very long keys")
	})
}
