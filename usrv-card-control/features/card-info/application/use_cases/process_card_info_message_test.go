package use_cases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockCardInfoRepository struct {
	mock.Mock
}

func (m *MockCardInfoRepository) Save(ctx context.Context, cardInfo *entities.StoredCardInfo) error {
	args := m.Called(ctx, cardInfo)
	return args.Error(0)
}

func (m *MockCardInfoRepository) FindByExternalReferenceID(ctx context.Context, externalReferenceID string) (*entities.StoredCardInfo, error) {
	args := m.Called(ctx, externalReferenceID)
	return args.Get(0).(*entities.StoredCardInfo), args.Error(1)
}

func (m *MockCardInfoRepository) Delete(ctx context.Context, externalReferenceID string) error {
	args := m.Called(ctx, externalReferenceID)
	return args.Error(0)
}

func (m *MockCardInfoRepository) FindExpiredRecords(ctx context.Context, currentTime int64) ([]*entities.StoredCardInfo, error) {
	args := m.Called(ctx, currentTime)
	return args.Get(0).([]*entities.StoredCardInfo), args.Error(1)
}

type MockEncryptionService struct {
	mock.Mock
}

func (m *MockEncryptionService) EncryptCardData(cardData value_objects.CardData, merchantID string) (value_objects.EncryptedCardData, error) {
	args := m.Called(cardData, merchantID)
	return args.Get(0).(value_objects.EncryptedCardData), args.Error(1)
}

type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) ValidateCardInfoMessage(message *entities.PxpCardInfoMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockValidationService) ValidateMerchantAccess(merchantID string) error {
	args := m.Called(merchantID)
	return args.Error(0)
}

func (m *MockValidationService) ValidatePrivateCredential(privateCredentialID, merchantID string) error {
	args := m.Called(privateCredentialID, merchantID)
	return args.Error(0)
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

// Test scenarios
func TestProcessCardInfoMessageUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		SubMerchantCode:      "sub-merchant-001",
		IDAffiliation:        "affiliation-001",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	encryptedData := value_objects.EncryptedCardData{
		EncryptedPan:  "encrypted-pan-data",
		EncryptedDate: "encrypted-date-data",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "private-cred-456", "merchant-123").Return(nil)
	mockRepo.On("FindByExternalReferenceID", ctx, "ext-ref-123").Return(&entities.StoredCardInfo{}, dynamoerror.ErrItemNotFound)
	mockEncryption.On("EncryptCardData", validMessage.Card, "merchant-123").Return(encryptedData, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.StoredCardInfo")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "ext-ref-123", response.ExternalReferenceID)
	assert.True(t, response.Success)
	assert.Greater(t, response.ProcessedAt, int64(0))

	mockRepo.AssertExpectations(t)
	mockEncryption.AssertExpectations(t)
	mockValidation.AssertExpectations(t)
}

func TestProcessCardInfoMessageUseCase_Execute_InvalidJSON(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: "invalid-json-data",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to parse SQS message")
}

func TestProcessCardInfoMessageUseCase_Execute_ValidationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID: "ext-ref-123",
		MerchantID:          "merchant-123",
		// Missing required fields to trigger validation error
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(errors.New("validation failed"))

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "message validation failed")
}

func TestProcessCardInfoMessageUseCase_Execute_MerchantAccessDenied(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(errors.New("access denied"))

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "merchant access validation failed")
}

func TestProcessCardInfoMessageUseCase_Execute_InvalidPrivateCredential(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "invalid-cred",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "invalid-cred", "merchant-123").Return(errors.New("invalid credential"))

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid private credential")
}

func TestProcessCardInfoMessageUseCase_Execute_AlreadyProcessed(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	existingCardInfo := &entities.StoredCardInfo{
		ExternalReferenceID: "ext-ref-123",
		MerchantID:          "merchant-123",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "private-cred-456", "merchant-123").Return(nil)
	mockRepo.On("FindByExternalReferenceID", ctx, "ext-ref-123").Return(existingCardInfo, nil)

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "ext-ref-123", response.ExternalReferenceID)
	assert.True(t, response.Success)
	assert.Greater(t, response.ProcessedAt, int64(0))

	// Verify that encryption and save were not called (idempotency)
	mockEncryption.AssertNotCalled(t, "EncryptCardData")
	mockRepo.AssertNotCalled(t, "Save")
}

func TestProcessCardInfoMessageUseCase_Execute_ExistenceCheckError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	// Setup mocks - repository check fails with unexpected error (NOT a "not found" error)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "private-cred-456", "merchant-123").Return(nil)

	// This simulates a database connection error or other unexpected repository failure
	// This should NOT be a "not found" error, but a real error like connection failure
	databaseError := errors.New("database connection failed")
	mockRepo.On("FindByExternalReferenceID", ctx, "ext-ref-123").Return((*entities.StoredCardInfo)(nil), databaseError)

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to check if already processed")
	assert.Contains(t, err.Error(), "database connection failed")

	// Verify that encryption and save were not called due to early failure
	mockEncryption.AssertNotCalled(t, "EncryptCardData")
	mockRepo.AssertNotCalled(t, "Save")

	// Verify the error was logged with the correct tag
	mockLogger.AssertCalled(t, "Error", "ProcessCardInfoMessage | ExistenceCheckError", databaseError)
}

func TestProcessCardInfoMessageUseCase_Execute_EncryptionError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "private-cred-456", "merchant-123").Return(nil)
	mockRepo.On("FindByExternalReferenceID", ctx, "ext-ref-123").Return(&entities.StoredCardInfo{}, dynamoerror.ErrItemNotFound)
	mockEncryption.On("EncryptCardData", validMessage.Card, "merchant-123").Return(value_objects.EncryptedCardData{}, errors.New("encryption failed"))

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to encrypt card data")
}

func TestProcessCardInfoMessageUseCase_Execute_SaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	validMessage := entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
		Card: value_objects.CardData{
			Pan:  "4111111111111111",
			Date: "1225",
		},
	}

	validMessageJSON, _ := json.Marshal(validMessage)
	request := ProcessCardInfoMessageRequest{
		SQSMessageBody: string(validMessageJSON),
	}

	encryptedData := value_objects.EncryptedCardData{
		EncryptedPan:  "encrypted-pan-data",
		EncryptedDate: "encrypted-date-data",
	}

	// Setup mocks - "not found" should be treated as normal case (continue processing)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockValidation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
	mockValidation.On("ValidateMerchantAccess", "merchant-123").Return(nil)
	mockValidation.On("ValidatePrivateCredential", "private-cred-456", "merchant-123").Return(nil)

	// For the existence check, we want "not found" behavior - your checkIfAlreadyProcessed should return (false, nil)
	// when it encounters ErrItemNotFound, so let's simulate that directly
	mockRepo.On("FindByExternalReferenceID", ctx, "ext-ref-123").Return((*entities.StoredCardInfo)(nil), dynamoerror.ErrItemNotFound)
	mockEncryption.On("EncryptCardData", validMessage.Card, "merchant-123").Return(encryptedData, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.StoredCardInfo")).Return(errors.New("save failed"))

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to save card info")
}

// Test helper to verify StoredCardInfo creation
func TestProcessCardInfoMessageUseCase_CreateStoredCardInfo_Validation(t *testing.T) {
	// Arrange
	mockRepo := &MockCardInfoRepository{}
	mockEncryption := &MockEncryptionService{}
	mockValidation := &MockValidationService{}
	mockLogger := &MockLogger{}

	useCase := NewProcessCardInfoMessageUseCase(mockRepo, mockEncryption, mockValidation, mockLogger)

	message := &entities.PxpCardInfoMessage{
		ExternalReferenceID:  "ext-ref-123",
		TransactionReference: "txn-ref-456",
		CardBrand:            "VISA",
		TerminalID:           "terminal-001",
		TransactionType:      "charge",
		TransactionStatus:    "APPROVAL",
		SubMerchantCode:      "sub-merchant-001",
		IDAffiliation:        "affiliation-001",
		MerchantID:           "merchant-123",
		PrivateCredentialID:  "private-cred-456",
	}

	encryptedData := value_objects.EncryptedCardData{
		EncryptedPan:  "encrypted-pan-data",
		EncryptedDate: "encrypted-date-data",
	}

	// Act
	storedCardInfo := useCase.createStoredCardInfo(message, encryptedData)

	// Assert
	assert.Equal(t, message.ExternalReferenceID, storedCardInfo.ExternalReferenceID)
	assert.Equal(t, message.TransactionReference, storedCardInfo.TransactionReference)
	assert.Equal(t, message.CardBrand, storedCardInfo.CardBrand)
	assert.Equal(t, message.TerminalID, storedCardInfo.TerminalID)
	assert.Equal(t, message.TransactionType, storedCardInfo.TransactionType)
	assert.Equal(t, message.TransactionStatus, storedCardInfo.TransactionStatus)
	assert.Equal(t, message.SubMerchantCode, storedCardInfo.SubMerchantCode)
	assert.Equal(t, message.IDAffiliation, storedCardInfo.IDAffiliation)
	assert.Equal(t, message.MerchantID, storedCardInfo.MerchantID)
	assert.Equal(t, message.PrivateCredentialID, storedCardInfo.PrivateCredentialID)
	assert.Equal(t, encryptedData, storedCardInfo.EncryptedCard)

	// Verify timestamps
	assert.Greater(t, storedCardInfo.CreatedAt, int64(0))
	assert.Greater(t, storedCardInfo.ExpiresAt, storedCardInfo.CreatedAt)
	assert.Greater(t, storedCardInfo.TransactionDate, int64(0))

	// Verify 180-day expiration (approximately)
	expectedExpiration := time.Now().AddDate(0, 0, 180).UnixMilli()
	timeDiff := storedCardInfo.ExpiresAt - expectedExpiration
	assert.True(t, timeDiff > -1000 && timeDiff < 1000, "Expiration should be approximately 180 days from now")
}
