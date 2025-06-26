package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
	"testing"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMerchantKeyProvider - mock for the key provider interface
type MockMerchantKeyProvider struct {
	mock.Mock
}

func (m *MockMerchantKeyProvider) GetMerchantPublicKey(merchantID string) (string, error) {
	args := m.Called(merchantID)
	return args.String(0), args.Error(1)
}

func (m *MockMerchantKeyProvider) HasMerchantKey(merchantID string) bool {
	args := m.Called(merchantID)
	return args.Bool(0)
}

// MockRSALogger - dedicated mock logger for RSA encryption service tests
type MockRSALogger struct {
	mock.Mock
}

func (m *MockRSALogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockRSALogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockRSALogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockRSALogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Test helper functions
func setupRSAEncryptionService(t *testing.T) (*RSAEncryptionService, *MockMerchantKeyProvider, *MockRSALogger) {
	t.Helper()
	mockKeyProvider := &MockMerchantKeyProvider{}
	mockLogger := &MockRSALogger{}
	service := NewRSAEncryptionService(mockKeyProvider, mockLogger).(*RSAEncryptionService)
	return service, mockKeyProvider, mockLogger
}

// Helper to generate a test RSA key pair
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	// Generate a test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err, "Failed to generate test RSA key")

	// Extract public key and encode as PEM
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err, "Failed to marshal public key")

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return privateKey, string(publicKeyPEM)
}

// Helper to create valid card data
func createValidCardData() value_objects.CardData {
	return value_objects.CardData{
		Pan:  "4111111111111111",
		Date: "1225",
	}
}

// Test EncryptCardData - Success Cases
func TestRSAEncryptionService_EncryptCardData_Success(t *testing.T) {
	t.Run("Successfully encrypt card data with valid key", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, encryptedData.EncryptedPan, "Encrypted PAN should not be empty")
		assert.NotEmpty(t, encryptedData.EncryptedDate, "Encrypted date should not be empty")

		// Verify base64 encoding
		_, err = base64.StdEncoding.DecodeString(encryptedData.EncryptedPan)
		assert.NoError(t, err, "Encrypted PAN should be valid base64")

		_, err = base64.StdEncoding.DecodeString(encryptedData.EncryptedDate)
		assert.NoError(t, err, "Encrypted date should be valid base64")

		// Verify encrypted data is different from original
		assert.NotEqual(t, cardData.Pan, encryptedData.EncryptedPan)
		assert.NotEqual(t, cardData.Date, encryptedData.EncryptedDate)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Encrypt different card data produces different results", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		merchantID := "MERCHANT123"

		cardData1 := value_objects.CardData{Pan: "4111111111111111", Date: "1225"}
		cardData2 := value_objects.CardData{Pan: "5555555555554444", Date: "0630"}

		// Setup mocks for both calls
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil).Times(2)

		// Act
		encrypted1, err1 := service.EncryptCardData(cardData1, merchantID)
		encrypted2, err2 := service.EncryptCardData(cardData2, merchantID)

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Different input should produce different encrypted output
		assert.NotEqual(t, encrypted1.EncryptedPan, encrypted2.EncryptedPan)
		assert.NotEqual(t, encrypted1.EncryptedDate, encrypted2.EncryptedDate)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test EncryptCardData - Failure Cases
func TestRSAEncryptionService_EncryptCardData_Failures(t *testing.T) {
	t.Run("Key retrieval error", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "UNKNOWN_MERCHANT"
		keyError := errors.New("key not found")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return("", keyError)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get public key for merchant UNKNOWN_MERCHANT")
		assert.Contains(t, err.Error(), "key not found")
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Invalid PEM format", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"
		invalidPEM := "this-is-not-a-valid-pem-key"

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(invalidPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse public key")
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Invalid PEM block", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"
		invalidPEMBlock := `-----BEGIN PUBLIC KEY-----
invalid-base64-content-that-cannot-be-decoded
-----END PUBLIC KEY-----`

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(invalidPEMBlock, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse public key")
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Non-RSA public key", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"

		// Create a non-RSA key (this is a mock - in real scenarios this could be ECDSA, etc.)
		nonRSAKey := `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEMKBCTNIcKUSDii11ySs3526iDZ8A
iTo7Tu6KPAqv7D7gS2XpJFbZiItSs3m9+9Ue6GnvHw/A6egHQ0XWChOJkw==
-----END PUBLIC KEY-----`

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(nonRSAKey, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse public key")
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Empty PEM content", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return("", nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse public key")
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)

		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test ValidateMerchantKey
func TestRSAEncryptionService_ValidateMerchantKey(t *testing.T) {
	t.Run("Merchant has valid key", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		merchantID := "MERCHANT123"

		// Setup mocks
		mockKeyProvider.On("HasMerchantKey", merchantID).Return(true)

		// Act
		err := service.ValidateMerchantKey(merchantID)

		// Assert
		assert.NoError(t, err)
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Merchant has no key", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		merchantID := "NO_KEY_MERCHANT"

		// Setup mocks
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("HasMerchantKey", merchantID).Return(false)

		// Act
		err := service.ValidateMerchantKey(merchantID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no public key registered for merchant: NO_KEY_MERCHANT")
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test parsePublicKey method (indirectly through EncryptCardData)
func TestRSAEncryptionService_ParsePublicKey(t *testing.T) {
	t.Run("Valid RSA public key formats", func(t *testing.T) {
		// Test different valid PEM formats
		testCases := []struct {
			name         string
			keyGenerator func(t *testing.T) string
		}{
			{
				name: "Standard 2048-bit RSA key",
				keyGenerator: func(t *testing.T) string {
					_, publicKeyPEM := generateTestKeyPair(t)
					return publicKeyPEM
				},
			},
			{
				name: "RSA key with Windows line endings",
				keyGenerator: func(t *testing.T) string {
					_, publicKeyPEM := generateTestKeyPair(t)
					// Convert to Windows line endings
					return strings.ReplaceAll(publicKeyPEM, "\n", "\r\n")
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
				cardData := createValidCardData()
				merchantID := "MERCHANT123"
				publicKeyPEM := tc.keyGenerator(t)

				// Setup mocks
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
				mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

				// Act
				encryptedData, err := service.EncryptCardData(cardData, merchantID)

				// Assert
				assert.NoError(t, err, "Should successfully parse %s", tc.name)
				assert.NotEmpty(t, encryptedData.EncryptedPan)
				assert.NotEmpty(t, encryptedData.EncryptedDate)
			})
		}
	})
}

// Test encryptData method (indirectly through EncryptCardData)
func TestRSAEncryptionService_EncryptData(t *testing.T) {
	t.Run("Encrypt various data sizes", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		merchantID := "MERCHANT123"

		testCases := []struct {
			name string
			data value_objects.CardData
		}{
			{
				name: "Standard card data",
				data: value_objects.CardData{Pan: "4111111111111111", Date: "1225"},
			},
			{
				name: "Short PAN",
				data: value_objects.CardData{Pan: "4111111111111", Date: "1225"},
			},
			{
				name: "Long PAN",
				data: value_objects.CardData{Pan: "4111111111111111111", Date: "1225"},
			},
			{
				name: "Different date format",
				data: value_objects.CardData{Pan: "4111111111111111", Date: "12/25"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup mocks
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
				mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

				// Act
				encryptedData, err := service.EncryptCardData(tc.data, merchantID)

				// Assert
				assert.NoError(t, err)
				assert.NotEmpty(t, encryptedData.EncryptedPan)
				assert.NotEmpty(t, encryptedData.EncryptedDate)

				// Verify different from original
				assert.NotEqual(t, tc.data.Pan, encryptedData.EncryptedPan)
				assert.NotEqual(t, tc.data.Date, encryptedData.EncryptedDate)
			})
		}
	})

	t.Run("Empty data encryption", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		merchantID := "MERCHANT123"
		emptyCardData := value_objects.CardData{Pan: "", Date: ""}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(emptyCardData, merchantID)

		// Assert
		assert.NoError(t, err, "Should be able to encrypt empty strings")

		// Even empty strings should produce non-empty encrypted output due to RSA padding
		assert.NotEmpty(t, encryptedData.EncryptedPan)
		assert.NotEmpty(t, encryptedData.EncryptedDate)
	})
}

// Test Logging Behavior
func TestRSAEncryptionService_LoggingBehavior(t *testing.T) {
	t.Run("Logs encryption start and success", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		cardData := createValidCardData()
		merchantID := "MERCHANT123"

		// Setup mocks
		mockLogger.On("Info", "RSAEncryptionService.EncryptCardData | Starting", "MerchantID: MERCHANT123").Return()
		mockLogger.On("Info", "RSAEncryptionService.EncryptCardData | Success", "MerchantID: MERCHANT123").Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, encryptedData.EncryptedPan)
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs key retrieval error", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "ERROR_MERCHANT"
		keyError := errors.New("key service error")

		// Setup mocks
		mockLogger.On("Info", "RSAEncryptionService.EncryptCardData | Starting", "MerchantID: ERROR_MERCHANT").Return()
		mockLogger.On("Error", "RSAEncryptionService.EncryptCardData | KeyRetrievalError", keyError).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return("", keyError)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs key parsing error", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		cardData := createValidCardData()
		merchantID := "INVALID_KEY_MERCHANT"
		invalidKey := "invalid-key"

		// Setup mocks
		mockLogger.On("Info", "RSAEncryptionService.EncryptCardData | Starting", "MerchantID: INVALID_KEY_MERCHANT").Return()
		mockLogger.On("Error", "RSAEncryptionService.EncryptCardData | KeyParsingError", mock.AnythingOfType("*errors.errorString")).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(invalidKey, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, merchantID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, value_objects.EncryptedCardData{}, encryptedData)
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs ValidateMerchantKey error", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		merchantID := "NO_KEY_MERCHANT"

		// Setup mocks
		mockLogger.On("Error", "RSAEncryptionService.ValidateMerchantKey | KeyNotFound", "MerchantID: NO_KEY_MERCHANT").Return()
		mockKeyProvider.On("HasMerchantKey", merchantID).Return(false)

		// Act
		err := service.ValidateMerchantKey(merchantID)

		// Assert
		assert.Error(t, err)
		mockKeyProvider.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Edge Cases
func TestRSAEncryptionService_EdgeCases(t *testing.T) {
	t.Run("Very long merchant ID", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		cardData := createValidCardData()
		longMerchantID := strings.Repeat("VERYLONGMERCHANTID", 10)

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", longMerchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(cardData, longMerchantID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, encryptedData.EncryptedPan)
		assert.NotEmpty(t, encryptedData.EncryptedDate)
	})

	t.Run("Special characters in card data", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		merchantID := "MERCHANT123"
		specialCardData := value_objects.CardData{
			Pan:  "4111-1111-1111-1111", // With dashes
			Date: "12/25",               // With slash
		}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(specialCardData, merchantID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, encryptedData.EncryptedPan)
		assert.NotEmpty(t, encryptedData.EncryptedDate)
	})

	t.Run("Unicode characters in card data", func(t *testing.T) {
		// Arrange
		service, mockKeyProvider, mockLogger := setupRSAEncryptionService(t)
		_, publicKeyPEM := generateTestKeyPair(t)
		merchantID := "MERCHANT123"
		unicodeCardData := value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225🔒", // With emoji
		}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockKeyProvider.On("GetMerchantPublicKey", merchantID).Return(publicKeyPEM, nil)

		// Act
		encryptedData, err := service.EncryptCardData(unicodeCardData, merchantID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, encryptedData.EncryptedPan)
		assert.NotEmpty(t, encryptedData.EncryptedDate)
	})
}
