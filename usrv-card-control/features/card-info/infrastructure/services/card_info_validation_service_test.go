package services

import (
	"testing"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockMerchantAccessProvider struct {
	mock.Mock
}

func (m *MockMerchantAccessProvider) HasCardInfoAccess(merchantID string) bool {
	args := m.Called(merchantID)
	return args.Bool(0)
}

func (m *MockMerchantAccessProvider) IsActiveMerchant(merchantID string) bool {
	args := m.Called(merchantID)
	return args.Bool(0)
}

type MockCredentialProvider struct {
	mock.Mock
}

func (m *MockCredentialProvider) ValidatePrivateCredential(privateCredentialID, merchantID string) bool {
	args := m.Called(privateCredentialID, merchantID)
	return args.Bool(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockLogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockLogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockLogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Helper function to create a valid card info message
func createValidCardInfoMessage() *entities.PxpCardInfoMessage {
	return &entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		SubMerchantCode:      "sub-merchant-001",
		IDAffiliation:        "affiliation-001",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}
}

// Test ValidateCardInfoMessage - Success Cases
func TestCardInfoValidationService_ValidateCardInfoMessage_Success(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
	message := createValidCardInfoMessage()

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

	// Act
	err := service.ValidateCardInfoMessage(message)

	// Assert
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

// Test ValidateCardInfoMessage - Required Field Validation
func TestCardInfoValidationService_ValidateCardInfoMessage_RequiredFields(t *testing.T) {
	testCases := []struct {
		name          string
		setupMessage  func() *entities.PxpCardInfoMessage
		expectedError string
	}{
		{
			name: "Missing ExternalReferenceID",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.ExternalReferenceID = ""
				return msg
			},
			expectedError: "externalReferenceId is required",
		},
		{
			name: "Missing TransactionReference",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.TransactionReference = ""
				return msg
			},
			expectedError: "transactionReference is required",
		},
		{
			name: "Missing MerchantID",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.MerchantID = ""
				return msg
			},
			expectedError: "merchant_id is required",
		},
		{
			name: "Missing PrivateCredentialID",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.PrivateCredentialID = ""
				return msg
			},
			expectedError: "privateCredentialId is required",
		},
		{
			name: "Missing Card PAN",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.Card.Pan = ""
				return msg
			},
			expectedError: "card.pan is required",
		},
		{
			name: "Missing Card Date",
			setupMessage: func() *entities.PxpCardInfoMessage {
				msg := createValidCardInfoMessage()
				msg.Card.Date = ""
				return msg
			},
			expectedError: "card.date is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockMerchantAccess := &MockMerchantAccessProvider{}
			mockCredentials := &MockCredentialProvider{}
			mockLogger := &MockLogger{}

			service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
			message := tc.setupMessage()

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			err := service.ValidateCardInfoMessage(message)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// Test ValidateCardInfoMessage - PAN Validation
func TestCardInfoValidationService_ValidateCardInfoMessage_PANValidation(t *testing.T) {
	testCases := []struct {
		name          string
		pan           string
		expectedError string
	}{
		{
			name:          "Valid PAN - 16 digits",
			pan:           "4111111111111111",
			expectedError: "",
		},
		{
			name:          "Valid PAN - 15 digits (AMEX)",
			pan:           "411111111111111",
			expectedError: "",
		},
		{
			name:          "Valid PAN - 13 digits",
			pan:           "4111111111111",
			expectedError: "",
		},
		{
			name:          "Invalid PAN - Too short (12 digits)",
			pan:           "411111111111",
			expectedError: "invalid PAN length: must be 13-19 digits",
		},
		{
			name:          "Invalid PAN - Too long (20 digits)",
			pan:           "41111111111111111111",
			expectedError: "invalid PAN length: must be 13-19 digits",
		},
		{
			name:          "Invalid PAN - Contains letters",
			pan:           "411111111111111a",
			expectedError: "invalid PAN format: must contain only digits",
		},
		{
			name:          "Invalid PAN - Contains spaces",
			pan:           "4111 1111 1111 1111",
			expectedError: "", // Should be valid after cleaning
		},
		{
			name:          "Invalid PAN - Contains dashes",
			pan:           "4111-1111-1111-1111",
			expectedError: "", // Should be valid after cleaning
		},
		{
			name:          "Invalid PAN - Special characters",
			pan:           "4111@1111#1111$1111",
			expectedError: "invalid PAN format: must contain only digits",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockMerchantAccess := &MockMerchantAccessProvider{}
			mockCredentials := &MockCredentialProvider{}
			mockLogger := &MockLogger{}

			service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
			message := createValidCardInfoMessage()
			message.Card.Pan = tc.pan

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			err := service.ValidateCardInfoMessage(message)

			// Assert
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Test ValidateCardInfoMessage - Expiration Date Validation
func TestCardInfoValidationService_ValidateCardInfoMessage_ExpirationDateValidation(t *testing.T) {
	testCases := []struct {
		name          string
		date          string
		expectedError string
	}{
		{
			name:          "Valid date - MMYY format",
			date:          "1225",
			expectedError: "",
		},
		{
			name:          "Valid date - MM/YY format",
			date:          "12/25",
			expectedError: "",
		},
		{
			name:          "Invalid date - Too short",
			date:          "125",
			expectedError: "invalid expiration date format: expected MMYY or MM/YY",
		},
		{
			name:          "Invalid date - Too long",
			date:          "12255",
			expectedError: "invalid expiration date format: expected MMYY or MM/YY",
		},
		{
			name:          "Invalid date - Contains letters",
			date:          "12ab",
			expectedError: "invalid expiration date format: must contain only digits",
		},
		{
			name:          "Invalid date - Special characters",
			date:          "12@5",
			expectedError: "invalid expiration date format: must contain only digits",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockMerchantAccess := &MockMerchantAccessProvider{}
			mockCredentials := &MockCredentialProvider{}
			mockLogger := &MockLogger{}

			service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
			message := createValidCardInfoMessage()
			message.Card.Date = tc.date

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			err := service.ValidateCardInfoMessage(message)

			// Assert
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Test ValidateCardInfoMessage - Card Brand Validation
func TestCardInfoValidationService_ValidateCardInfoMessage_CardBrandValidation(t *testing.T) {
	testCases := []struct {
		name          string
		cardBrand     string
		expectedError string
	}{
		{
			name:          "Valid brand - VISA",
			cardBrand:     "VISA",
			expectedError: "",
		},
		{
			name:          "Valid brand - MASTERCARD",
			cardBrand:     "MASTERCARD",
			expectedError: "",
		},
		{
			name:          "Valid brand - AMEX",
			cardBrand:     "AMEX",
			expectedError: "",
		},
		{
			name:          "Valid brand - lowercase visa",
			cardBrand:     "visa",
			expectedError: "",
		},
		{
			name:          "Valid brand - mixed case",
			cardBrand:     "MasterCard",
			expectedError: "",
		},
		{
			name:          "Invalid brand",
			cardBrand:     "UNKNOWN",
			expectedError: "invalid card brand: UNKNOWN",
		},
		{
			name:          "Empty brand",
			cardBrand:     "",
			expectedError: "invalid card brand:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockMerchantAccess := &MockMerchantAccessProvider{}
			mockCredentials := &MockCredentialProvider{}
			mockLogger := &MockLogger{}

			service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
			message := createValidCardInfoMessage()
			message.CardBrand = tc.cardBrand

			// Setup mocks
			mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

			// Act
			err := service.ValidateCardInfoMessage(message)

			// Assert
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Test ValidateMerchantAccess - Success
func TestCardInfoValidationService_ValidateMerchantAccess_Success(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockMerchantAccess.On("IsActiveMerchant", "merchant-123").Return(true)
	mockMerchantAccess.On("HasCardInfoAccess", "merchant-123").Return(true)

	// Act
	err := service.ValidateMerchantAccess("merchant-123")

	// Assert
	assert.NoError(t, err)
	mockMerchantAccess.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test ValidateMerchantAccess - Inactive Merchant
func TestCardInfoValidationService_ValidateMerchantAccess_InactiveMerchant(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockMerchantAccess.On("IsActiveMerchant", "merchant-123").Return(false)

	// Act
	err := service.ValidateMerchantAccess("merchant-123")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant is not active: merchant-123")
	mockMerchantAccess.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test ValidateMerchantAccess - No Card Info Access
func TestCardInfoValidationService_ValidateMerchantAccess_NoCardInfoAccess(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockMerchantAccess.On("IsActiveMerchant", "merchant-123").Return(true)
	mockMerchantAccess.On("HasCardInfoAccess", "merchant-123").Return(false)

	// Act
	err := service.ValidateMerchantAccess("merchant-123")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant does not have card info access: merchant-123")
	mockMerchantAccess.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test ValidatePrivateCredential - Success
func TestCardInfoValidationService_ValidatePrivateCredential_Success(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockCredentials.On("ValidatePrivateCredential", "private-cred-123", "merchant-123").Return(true)

	// Act
	err := service.ValidatePrivateCredential("private-cred-123", "merchant-123")

	// Assert
	assert.NoError(t, err)
	mockCredentials.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test ValidatePrivateCredential - Invalid Credential
func TestCardInfoValidationService_ValidatePrivateCredential_InvalidCredential(t *testing.T) {
	// Arrange
	mockMerchantAccess := &MockMerchantAccessProvider{}
	mockCredentials := &MockCredentialProvider{}
	mockLogger := &MockLogger{}

	service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockCredentials.On("ValidatePrivateCredential", "invalid-cred", "merchant-123").Return(false)

	// Act
	err := service.ValidatePrivateCredential("invalid-cred", "merchant-123")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private credential for merchant: merchant-123")
	mockCredentials.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test Edge Cases
func TestCardInfoValidationService_EdgeCases(t *testing.T) {
	t.Run("PAN with mixed spaces and dashes", func(t *testing.T) {
		// Arrange
		mockMerchantAccess := &MockMerchantAccessProvider{}
		mockCredentials := &MockCredentialProvider{}
		mockLogger := &MockLogger{}

		service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
		message := createValidCardInfoMessage()
		message.Card.Pan = "4111 1111-1111 1111" // Mixed formatting

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		err := service.ValidateCardInfoMessage(message)

		// Assert
		assert.NoError(t, err, "Should handle mixed PAN formatting")
	})

	t.Run("All supported card brands", func(t *testing.T) {
		supportedBrands := []string{"VISA", "MASTERCARD", "AMEX", "DISCOVER", "DINERS", "JCB"}

		for _, brand := range supportedBrands {
			t.Run(brand, func(t *testing.T) {
				// Arrange
				mockMerchantAccess := &MockMerchantAccessProvider{}
				mockCredentials := &MockCredentialProvider{}
				mockLogger := &MockLogger{}

				service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
				message := createValidCardInfoMessage()
				message.CardBrand = brand

				// Setup mocks
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

				// Act
				err := service.ValidateCardInfoMessage(message)

				// Assert
				assert.NoError(t, err, "Should support brand: %s", brand)
			})
		}
	})

	t.Run("Maximum length PAN", func(t *testing.T) {
		// Arrange
		mockMerchantAccess := &MockMerchantAccessProvider{}
		mockCredentials := &MockCredentialProvider{}
		mockLogger := &MockLogger{}

		service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
		message := createValidCardInfoMessage()
		message.Card.Pan = "1234567890123456789" // 19 digits (max allowed)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		err := service.ValidateCardInfoMessage(message)

		// Assert
		assert.NoError(t, err, "Should accept 19-digit PAN")
	})

	t.Run("Minimum length PAN", func(t *testing.T) {
		// Arrange
		mockMerchantAccess := &MockMerchantAccessProvider{}
		mockCredentials := &MockCredentialProvider{}
		mockLogger := &MockLogger{}

		service := NewCardInfoValidationService(mockMerchantAccess, mockCredentials, mockLogger)
		message := createValidCardInfoMessage()
		message.Card.Pan = "1234567890123" // 13 digits (min allowed)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Act
		err := service.ValidateCardInfoMessage(message)

		// Assert
		assert.NoError(t, err, "Should accept 13-digit PAN")
	})
}
