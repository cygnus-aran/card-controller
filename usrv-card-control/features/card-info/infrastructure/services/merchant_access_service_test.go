package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMerchantAccessLogger - dedicated mock logger for merchant access service tests
type MockMerchantAccessLogger struct {
	mock.Mock
}

func (m *MockMerchantAccessLogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantAccessLogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantAccessLogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockMerchantAccessLogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Test helper functions
func setupMerchantAccessService(t *testing.T) (*MerchantAccessService, *MockMerchantAccessLogger) {
	t.Helper()
	mockLogger := &MockMerchantAccessLogger{}
	service := NewMerchantAccessService(mockLogger).(*MerchantAccessService)
	return service, mockLogger
}

// Helper to set and cleanup environment variables
func setMerchantAccessEnvVar(t *testing.T, key, value string) {
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

// Test HasCardInfoAccess - Success Cases
func TestMerchantAccessService_HasCardInfoAccess_Success(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Merchant in allowed list - single merchant",
			merchantID: "MERCHANT123",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123")
			},
			expectedResult: true,
			description:    "Should grant access when merchant is in allowed list",
		},
		{
			name:       "Merchant in allowed list - multiple merchants",
			merchantID: "MERCHANT456",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123,MERCHANT456,MERCHANT789")
			},
			expectedResult: true,
			description:    "Should grant access when merchant is in comma-separated list",
		},
		{
			name:       "Merchant in allowed list - with spaces",
			merchantID: "MERCHANT_SPACED",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", " MERCHANT123 , MERCHANT_SPACED , MERCHANT789 ")
			},
			expectedResult: true,
			description:    "Should handle spaces in environment variable correctly",
		},
		{
			name:       "Test environment - any merchant allowed",
			merchantID: "ANY_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "test")
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
			},
			expectedResult: true,
			description:    "Should allow any merchant in test environment",
		},
		{
			name:       "Dev environment - any merchant allowed",
			merchantID: "DEV_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "dev")
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
			},
			expectedResult: true,
			description:    "Should allow any merchant in dev environment",
		},
		{
			name:       "Local environment - any merchant allowed",
			merchantID: "LOCAL_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "local")
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
			},
			expectedResult: true,
			description:    "Should allow any merchant in local environment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantAccessService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.HasCardInfoAccess(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test HasCardInfoAccess - Failure Cases
func TestMerchantAccessService_HasCardInfoAccess_Failures(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Merchant not in allowed list",
			merchantID: "UNAUTHORIZED_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123,MERCHANT456")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should deny access when merchant not in allowed list",
		},
		{
			name:       "Empty allowed merchants in production",
			merchantID: "ANY_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should deny access when no merchants are explicitly allowed in production",
		},
		{
			name:       "Case sensitive merchant matching",
			merchantID: "merchant123",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should be case sensitive when matching merchant IDs",
		},
		{
			name:       "Partial merchant ID match",
			merchantID: "MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should require exact merchant ID match, not partial",
		},
		{
			name:       "UAT environment with restricted access",
			merchantID: "TEST_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "uat")
			},
			expectedResult: false,
			description:    "Should restrict access in UAT environment when no merchants allowed",
		},
		{
			name:       "Staging environment with restricted access",
			merchantID: "TEST_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "staging")
			},
			expectedResult: false,
			description:    "Should restrict access in staging environment when no merchants allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantAccessService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.HasCardInfoAccess(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test IsActiveMerchant - Success Cases
func TestMerchantAccessService_IsActiveMerchant_Success(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Active merchant - not in inactive list",
			merchantID: "ACTIVE_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "INACTIVE1,INACTIVE2")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: true,
			description:    "Should return true when merchant is not in inactive list",
		},
		{
			name:       "Active merchant - empty inactive list",
			merchantID: "ANY_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: true,
			description:    "Should return true when no merchants are marked inactive",
		},
		{
			name:       "Active merchant - test environment",
			merchantID: "TEST_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "test")
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "")
			},
			expectedResult: true,
			description:    "Should assume active in test environment",
		},
		{
			name:       "Active merchant - dev environment",
			merchantID: "DEV_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "dev")
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "INACTIVE_MERCHANT")
			},
			expectedResult: true,
			description:    "Should assume active in dev environment",
		},
		{
			name:       "Active merchant - local environment",
			merchantID: "LOCAL_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "USRV_STAGE", "local")
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "")
			},
			expectedResult: true,
			description:    "Should assume active in local environment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantAccessService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.IsActiveMerchant(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test IsActiveMerchant - Failure Cases
func TestMerchantAccessService_IsActiveMerchant_Failures(t *testing.T) {
	testCases := []struct {
		name           string
		merchantID     string
		setupEnv       func(t *testing.T)
		expectedResult bool
		description    string
	}{
		{
			name:       "Inactive merchant - in inactive list",
			merchantID: "INACTIVE_MERCHANT",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "INACTIVE_MERCHANT,ANOTHER_INACTIVE")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should return false when merchant is in inactive list",
		},
		{
			name:       "Inactive merchant - with spaces in list",
			merchantID: "SPACED_INACTIVE",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", " INACTIVE1 , SPACED_INACTIVE , INACTIVE2 ")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: false,
			description:    "Should handle spaces in inactive merchants list",
		},
		{
			name:       "Case sensitive inactive matching",
			merchantID: "inactive_merchant",
			setupEnv: func(t *testing.T) {
				setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "INACTIVE_MERCHANT")
				setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")
			},
			expectedResult: true,
			description:    "Should be case sensitive - lowercase should not match uppercase in inactive list",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantAccessService(t)
			tc.setupEnv(t)

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.IsActiveMerchant(tc.merchantID)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test Environment Detection in Access Control
func TestMerchantAccessService_EnvironmentDetection(t *testing.T) {
	testCases := []struct {
		name         string
		usrvStage    string
		expectAccess bool
		description  string
	}{
		{
			name:         "Test environment allows access",
			usrvStage:    "test",
			expectAccess: true,
			description:  "Test environment should allow access for any merchant",
		},
		{
			name:         "Dev environment allows access",
			usrvStage:    "dev",
			expectAccess: true,
			description:  "Dev environment should allow access for any merchant",
		},
		{
			name:         "Local environment allows access",
			usrvStage:    "local",
			expectAccess: true,
			description:  "Local environment should allow access for any merchant",
		},
		{
			name:         "Production environment restricts access",
			usrvStage:    "prod",
			expectAccess: false,
			description:  "Production environment should restrict access without explicit configuration",
		},
		{
			name:         "UAT environment restricts access",
			usrvStage:    "uat",
			expectAccess: false,
			description:  "UAT environment should restrict access without explicit configuration",
		},
		{
			name:         "Staging environment restricts access",
			usrvStage:    "staging",
			expectAccess: false,
			description:  "Staging environment should restrict access without explicit configuration",
		},
		{
			name:         "Unknown environment restricts access",
			usrvStage:    "unknown",
			expectAccess: false,
			description:  "Unknown environment should restrict access",
		},
		{
			name:         "Empty environment restricts access",
			usrvStage:    "",
			expectAccess: false,
			description:  "Empty environment should restrict access",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			service, mockLogger := setupMerchantAccessService(t)
			setMerchantAccessEnvVar(t, "USRV_STAGE", tc.usrvStage)
			setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "") // No explicit merchants allowed

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			result := service.HasCardInfoAccess("TEST_MERCHANT")

			// Assert
			assert.Equal(t, tc.expectAccess, result, tc.description)
		})
	}
}

// Test Logging Behavior
func TestMerchantAccessService_LoggingBehavior(t *testing.T) {
	t.Run("Logs HasCardInfoAccess start", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123")

		// Setup mocks
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | Starting", "MerchantID: MERCHANT123").Return()
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | AccessGranted", "MerchantID: MERCHANT123").Return()

		// Act
		result := service.HasCardInfoAccess("MERCHANT123")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs HasCardInfoAccess access denied", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "OTHER_MERCHANT")
		setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | Starting", "MerchantID: DENIED_MERCHANT").Return()
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | AccessDenied", "MerchantID: DENIED_MERCHANT").Return()

		// Act
		result := service.HasCardInfoAccess("DENIED_MERCHANT")

		// Assert
		assert.False(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs IsActiveMerchant start", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "")

		// Setup mocks
		mockLogger.On("Info", "MerchantAccessService.IsActiveMerchant | Starting", "MerchantID: ACTIVE_MERCHANT").Return()
		mockLogger.On("Info", "MerchantAccessService.IsActiveMerchant | AssumedActive", "MerchantID: ACTIVE_MERCHANT").Return()

		// Act
		result := service.IsActiveMerchant("ACTIVE_MERCHANT")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs IsActiveMerchant inactive", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "INACTIVE_MERCHANT")
		setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", "MerchantAccessService.IsActiveMerchant | Starting", "MerchantID: INACTIVE_MERCHANT").Return()
		mockLogger.On("Info", "MerchantAccessService.IsActiveMerchant | MerchantInactive", "MerchantID: INACTIVE_MERCHANT").Return()

		// Act
		result := service.IsActiveMerchant("INACTIVE_MERCHANT")

		// Assert
		assert.False(t, result)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs test environment access", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "USRV_STAGE", "test")

		// Setup mocks
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | Starting", "MerchantID: ANY_MERCHANT").Return()
		mockLogger.On("Info", "MerchantAccessService.HasCardInfoAccess | TestEnvironmentAccess", "MerchantID: ANY_MERCHANT").Return()

		// Act
		result := service.HasCardInfoAccess("ANY_MERCHANT")

		// Assert
		assert.True(t, result)
		mockLogger.AssertExpectations(t)
	})
}

// Test Edge Cases
func TestMerchantAccessService_EdgeCases(t *testing.T) {
	t.Run("Single merchant in allowed list", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "SINGLE_MERCHANT")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.HasCardInfoAccess("SINGLE_MERCHANT")

		// Assert
		assert.True(t, result)
	})

	t.Run("Single merchant in inactive list", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_INACTIVE_MERCHANTS", "SINGLE_INACTIVE")
		setMerchantAccessEnvVar(t, "USRV_STAGE", "prod")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.IsActiveMerchant("SINGLE_INACTIVE")

		// Assert
		assert.False(t, result)
	})

	t.Run("Empty merchant ID", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", "MERCHANT123")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		accessResult := service.HasCardInfoAccess("")
		activeResult := service.IsActiveMerchant("")

		// Assert
		assert.False(t, accessResult, "Empty merchant ID should not have access")
		assert.True(t, activeResult, "Empty merchant ID should be considered active by default")
	})

	t.Run("Very long merchant ID", func(t *testing.T) {
		// Arrange
		service, mockLogger := setupMerchantAccessService(t)
		longMerchantID := "VERY_LONG_MERCHANT_ID_THAT_EXCEEDS_NORMAL_LENGTH_EXPECTATIONS_AND_CONTINUES_FOR_A_WHILE"
		setMerchantAccessEnvVar(t, "CARD_INFO_ALLOWED_MERCHANTS", longMerchantID)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		result := service.HasCardInfoAccess(longMerchantID)

		// Assert
		assert.True(t, result, "Should handle very long merchant IDs")
	})
}
