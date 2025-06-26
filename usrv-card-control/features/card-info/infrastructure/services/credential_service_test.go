package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCredentialLogger - dedicated mock logger for credential service tests
type MockCredentialLogger struct {
	mock.Mock
}

func (m *MockCredentialLogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockCredentialLogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockCredentialLogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockCredentialLogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Test helper functions
func setupCredentialService(t *testing.T) (*CredentialService, *MockCredentialLogger) {
	t.Helper()
	mockLogger := &MockCredentialLogger{}
	service := NewCredentialService(mockLogger).(*CredentialService)
	return service, mockLogger
}

// Helper to set and cleanup environment variables
func setEnvVar(t *testing.T, key, value string) {
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

// Test ValidatePrivateCredential - Success Cases
func TestCredentialService_ValidatePrivateCredential_Success(t *testing.T) {
	testCases := []struct {
		name                string
		privateCredentialID string
		merchantID          string
		setupEnv            func(t *testing.T)
		expectedResult      bool
	}{
		{
			name:                "Valid credential from environment variable",
			privateCredentialID: "valid-credential-123",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "valid-credential-123,another-credential-789")
			},
			expectedResult: true,
		},
		{
			name:                "Valid credential with spaces in env list",
			privateCredentialID: "spaced-credential",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", " spaced-credential , another-credential ")
			},
			expectedResult: true,
		},
		{
			name:                "Valid credential matching merchant pattern",
			privateCredentialID: "PRIV_MERCHANT123_CREDENTIAL",
			merchantID:          "MERCHANT123",
			setupEnv: func(t *testing.T) {
				// No env var set, should use pattern matching
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: true,
		},
		{
			name:                "Valid credential in test environment - TEST_ prefix",
			privateCredentialID: "TEST_CREDENTIAL_123",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "test")
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: true,
		},
		{
			name:                "Valid credential in dev environment - long credential",
			privateCredentialID: "this-is-a-very-long-credential-that-should-pass",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "dev")
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: true,
		},
		{
			name:                "Valid credential in local environment",
			privateCredentialID: "TEST_LOCAL_CREDENTIAL",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "local")
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupCredentialService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.ValidatePrivateCredential(tc.privateCredentialID, tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test ValidatePrivateCredential - Failure Cases
func TestCredentialService_ValidatePrivateCredential_Failures(t *testing.T) {
	testCases := []struct {
		name                string
		privateCredentialID string
		merchantID          string
		setupEnv            func(t *testing.T)
		expectedResult      bool
	}{
		{
			name:                "Empty credential ID",
			privateCredentialID: "",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
		},
		{
			name:                "Credential not in environment list",
			privateCredentialID: "invalid-credential",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "valid-credential-123,another-credential-789")
				setEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
		},
		{
			name:                "Wrong merchant pattern",
			privateCredentialID: "PRIV_WRONGMERCHANT_CREDENTIAL",
			merchantID:          "RIGHTMERCHANT",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
				setEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
		},
		{
			name:                "No pattern match in production",
			privateCredentialID: "random-credential-without-pattern",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
				setEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
		},
		{
			name:                "Short credential in test environment",
			privateCredentialID: "short",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "test")
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: false,
		},
		{
			name:                "No TEST_ prefix and short in test environment",
			privateCredentialID: "shortcred",
			merchantID:          "merchant-456",
			setupEnv: func(t *testing.T) {
				setEnvVar(t, "USRV_STAGE", "test")
				setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupCredentialService(t)
			tc.setupEnv(t)

			// Setup mocks
			if tc.privateCredentialID == "" {
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
				mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
			} else {
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
				mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
			}

			// Act
			result := service.ValidatePrivateCredential(tc.privateCredentialID, tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test Environment Detection Logic
func TestCredentialService_EnvironmentDetection(t *testing.T) {
	testCases := []struct {
		name        string
		usrvStage   string
		isTestEnv   bool
		description string
	}{
		{
			name:        "Development environment",
			usrvStage:   "dev",
			isTestEnv:   true,
			description: "Should be detected as test environment",
		},
		{
			name:        "Test environment",
			usrvStage:   "test",
			isTestEnv:   true,
			description: "Should be detected as test environment",
		},
		{
			name:        "Local environment",
			usrvStage:   "local",
			isTestEnv:   true,
			description: "Should be detected as test environment",
		},
		{
			name:        "Production environment",
			usrvStage:   "prod",
			isTestEnv:   false,
			description: "Should NOT be detected as test environment",
		},
		{
			name:        "UAT environment",
			usrvStage:   "uat",
			isTestEnv:   false,
			description: "Should NOT be detected as test environment",
		},
		{
			name:        "Staging environment",
			usrvStage:   "staging",
			isTestEnv:   false,
			description: "Should NOT be detected as test environment",
		},
		{
			name:        "Empty environment",
			usrvStage:   "",
			isTestEnv:   false,
			description: "Should NOT be detected as test environment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			setEnvVar(t, "USRV_STAGE", tc.usrvStage)

			// Act
			result := isTestEnvironment()

			// Assert
			assert.Equal(t, tc.isTestEnv, result, tc.description)
		})
	}
}

// Test Merchant Pattern Matching
func TestCredentialService_MerchantPatternMatching(t *testing.T) {
	testCases := []struct {
		name                string
		privateCredentialID string
		merchantID          string
		shouldMatch         bool
	}{
		{
			name:                "Exact merchant pattern match",
			privateCredentialID: "PRIV_MERCHANT123_CREDENTIAL",
			merchantID:          "MERCHANT123",
			shouldMatch:         true,
		},
		{
			name:                "Pattern with additional suffix",
			privateCredentialID: "PRIV_TESTMERCHANT_PROD_KEY",
			merchantID:          "TESTMERCHANT",
			shouldMatch:         true,
		},
		{
			name:                "Pattern mismatch - wrong merchant",
			privateCredentialID: "PRIV_MERCHANT456_CREDENTIAL",
			merchantID:          "MERCHANT123",
			shouldMatch:         false,
		},
		{
			name:                "No PRIV_ prefix",
			privateCredentialID: "MERCHANT123_CREDENTIAL",
			merchantID:          "MERCHANT123",
			shouldMatch:         false,
		},
		{
			name:                "Case sensitive merchant ID",
			privateCredentialID: "PRIV_merchant123_CREDENTIAL",
			merchantID:          "MERCHANT123",
			shouldMatch:         false,
		},
		{
			name:                "Empty merchant ID",
			privateCredentialID: "PRIV__CREDENTIAL",
			merchantID:          "",
			shouldMatch:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupCredentialService(t)
			setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "") // No env credentials
			setEnvVar(t, "USRV_STAGE", "prod")              // Production environment

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
			if tc.shouldMatch {
				// Should not call Error if pattern matches
			} else {
				mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
			}

			// Act
			result := service.ValidatePrivateCredential(tc.privateCredentialID, tc.merchantID)

			// Assert
			assert.Equal(t, tc.shouldMatch, result)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test Logging Behavior
func TestCredentialService_LoggingBehavior(t *testing.T) {
	t.Run("Logs start of validation", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "valid-cred")

		// Setup mocks
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | Starting", "MerchantID: test-merchant").Return()
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | CredentialValid", "MerchantID: test-merchant").Return()

		// Act
		result := service.ValidatePrivateCredential("valid-cred", "test-merchant")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs error for empty credential", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)

		// Setup mocks
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | Starting", "MerchantID: test-merchant").Return()
		mockLogger.On("Error", "CredentialService.ValidatePrivateCredential | EmptyCredential", "MerchantID: test-merchant").Return()

		// Act
		result := service.ValidatePrivateCredential("", "test-merchant")

		// Assert
		assert.False(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs pattern match success", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
		setEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | Starting", "MerchantID: TESTMERCHANT").Return()
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | PatternMatched", "MerchantID: TESTMERCHANT").Return()

		// Act
		result := service.ValidatePrivateCredential("PRIV_TESTMERCHANT_KEY", "TESTMERCHANT")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs test environment validation", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
		setEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | Starting", "MerchantID: test-merchant").Return()
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | TestEnvironmentValid", "MerchantID: test-merchant").Return()

		// Act
		result := service.ValidatePrivateCredential("TEST_LONG_CREDENTIAL", "test-merchant")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs credential invalid", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
		setEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "CredentialService.ValidatePrivateCredential | Starting", "MerchantID: test-merchant").Return()
		mockLogger.On("Error", "CredentialService.ValidatePrivateCredential | CredentialInvalid", "MerchantID: test-merchant").Return()

		// Act
		result := service.ValidatePrivateCredential("invalid-credential", "test-merchant")

		// Assert
		assert.False(t, result)
		mockLogger.AssertExpectations(t)
	})
}

// Test Edge Cases and Complex Scenarios
func TestCredentialService_EdgeCases(t *testing.T) {
	t.Run("Environment variable with single credential", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "single-credential")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.ValidatePrivateCredential("single-credential", "merchant")

		// Assert
		assert.True(t, result)
	})

	t.Run("Environment variable with empty string", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "CARD_INFO_VALID_CREDENTIALS", "")
		setEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.ValidatePrivateCredential("any-credential", "merchant")

		// Assert
		assert.False(t, result)
	})

	t.Run("Very long credential in test environment", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "USRV_STAGE", "dev")
		longCredential := "this-is-a-very-long-credential-that-exceeds-ten-characters-easily"

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.ValidatePrivateCredential(longCredential, "merchant")

		// Assert
		assert.True(t, result)
	})

	t.Run("Exactly 10 character credential in test environment", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.ValidatePrivateCredential("exactly10c", "merchant") // Exactly 10 chars

		// Assert
		assert.False(t, result, "Should require MORE than 10 characters")
	})

	t.Run("11 character credential in test environment", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupCredentialService(t)
		setEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.ValidatePrivateCredential("exactly11ch", "merchant") // 11 chars

		// Assert
		assert.True(t, result, "Should accept more than 10 characters")
	})
}
