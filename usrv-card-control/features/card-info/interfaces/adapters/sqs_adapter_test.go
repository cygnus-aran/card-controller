// features/card-info/interfaces/adapters/sqs_adapter_test.go
package adapters

import (
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	"context"
	"errors"
	"testing"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/application/use_cases"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/interfaces"
	"bitbucket.org/kushki/usrv-card-control/mocks"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for use case dependencies
type MockCardInfoRepository struct {
	mock.Mock
}

func (m *MockCardInfoRepository) Save(ctx context.Context, cardInfo *entities.StoredCardInfo) error {
	args := m.Called(ctx, cardInfo)
	return args.Error(0)
}

func (m *MockCardInfoRepository) FindByExternalReferenceID(ctx context.Context, externalReferenceID string) (*entities.StoredCardInfo, error) {
	args := m.Called(ctx, externalReferenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StoredCardInfo), args.Error(1)
}

func (m *MockCardInfoRepository) Delete(ctx context.Context, externalReferenceID string) error {
	args := m.Called(ctx, externalReferenceID)
	return args.Error(0)
}

func (m *MockCardInfoRepository) FindExpiredRecords(ctx context.Context, currentTime int64) ([]*entities.StoredCardInfo, error) {
	args := m.Called(ctx, currentTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

func TestSQSAdapter_ProcessCardInfoMessage(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockCardInfoRepository, *MockEncryptionService, *MockValidationService)
		messageBody   string
		expectedError bool
		errorContains string
	}{
		{
			name: "should process card info message successfully",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// Validation should pass
				validation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
				validation.On("ValidateMerchantAccess", "MERCHANT_123").Return(nil)
				validation.On("ValidatePrivateCredential", "PRIV_CRED_123", "MERCHANT_123").Return(nil)

				// Repository should not find existing record (for idempotency check)
				// Looking at the use case code: if err != nil, return false, nil (meaning "not found, continue")
				// So we return an error to simulate "not found"
				repo.On("FindByExternalReferenceID", mock.Anything, "EXT_REF_123").Return((*entities.StoredCardInfo)(nil), errors.New("item not found"))

				// Encryption should work
				encryption.On("EncryptCardData", mock.AnythingOfType("value_objects.CardData"), "MERCHANT_123").Return(
					value_objects.EncryptedCardData{
						EncryptedPan:  "encrypted_pan_data",
						EncryptedDate: "encrypted_date_data",
					}, nil)

				// Repository save should work
				repo.On("Save", mock.Anything, mock.AnythingOfType("*entities.StoredCardInfo")).Return(nil)
			},
			messageBody: `{
				"card": {"pan": "4111111111111111", "date": "1225"},
				"externalReferenceId": "EXT_REF_123",
				"transactionReference": "TXN_REF_123",
				"card_brand": "VISA",
				"terminalId": "TERM_123",
				"transactionType": "charge",
				"transaction_status": "APPROVAL",
				"sub_merchant_code": "SUB_123",
				"id_affiliation": "AFF_123",
				"merchant_id": "MERCHANT_123",
				"privateCredentialId": "PRIV_CRED_123"
			}`,
			expectedError: false,
		},
		{
			name: "should return error when validation fails",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// Validation should fail
				validation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(errors.New("validation failed"))
			},
			messageBody:   `{"externalReferenceId":"test-123"}`,
			expectedError: true,
			errorContains: "message validation failed",
		},
		{
			name: "should return error when encryption fails",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// Validation should pass
				validation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
				validation.On("ValidateMerchantAccess", "MERCHANT_123").Return(nil)
				validation.On("ValidatePrivateCredential", "PRIV_CRED_123", "MERCHANT_123").Return(nil)

				// Repository should not find existing record
				repo.On("FindByExternalReferenceID", mock.Anything, "EXT_REF_123").Return((*entities.StoredCardInfo)(nil), dynamoerror.ErrItemNotFound)

				// Encryption should fail
				encryption.On("EncryptCardData", mock.AnythingOfType("value_objects.CardData"), "MERCHANT_123").Return(
					value_objects.EncryptedCardData{}, errors.New("encryption failed"))
			},
			messageBody: `{
				"card": {"pan": "4111111111111111", "date": "1225"},
				"externalReferenceId": "EXT_REF_123",
				"merchant_id": "MERCHANT_123",
				"privateCredentialId": "PRIV_CRED_123"
			}`,
			expectedError: true,
			errorContains: "failed to encrypt card data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockCardInfoRepository{}
			mockEncryption := &MockEncryptionService{}
			mockValidation := &MockValidationService{}
			mockLogger := mocks.GetMockLogger(t)

			tt.setupMocks(mockRepo, mockEncryption, mockValidation)

			// Create real use case with mocked dependencies
			useCase := use_cases.NewProcessCardInfoMessageUseCase(
				mockRepo,
				mockEncryption,
				mockValidation,
				mockLogger,
			)

			// Create adapter with real use case
			adapter := NewSQSAdapter(useCase, mockLogger)

			// Execute - using the interface method
			err := adapter.ProcessCardInfoMessage(context.Background(), tt.messageBody)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockEncryption.AssertExpectations(t)
			mockValidation.AssertExpectations(t)
		})
	}
}

func TestSQSAdapter_HandleSQSEvent(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockCardInfoRepository, *MockEncryptionService, *MockValidationService)
		event         events.SQSEvent
		expectedError bool
		errorContains string
	}{
		{
			name: "should process single SQS record successfully",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// Setup successful processing
				validation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(nil)
				validation.On("ValidateMerchantAccess", "MERCHANT_123").Return(nil)
				validation.On("ValidatePrivateCredential", "PRIV_CRED_123", "MERCHANT_123").Return(nil)
				repo.On("FindByExternalReferenceID", mock.Anything, "EXT_REF_123").Return((*entities.StoredCardInfo)(nil), dynamoerror.ErrItemNotFound)
				encryption.On("EncryptCardData", mock.AnythingOfType("value_objects.CardData"), "MERCHANT_123").Return(
					value_objects.EncryptedCardData{
						EncryptedPan:  "encrypted_pan_data",
						EncryptedDate: "encrypted_date_data",
					}, nil)
				repo.On("Save", mock.Anything, mock.AnythingOfType("*entities.StoredCardInfo")).Return(nil)
			},
			event: events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId: "message-123",
						Body: `{
							"card": {"pan": "4111111111111111", "date": "1225"},
							"externalReferenceId": "EXT_REF_123",
							"merchant_id": "MERCHANT_123",
							"privateCredentialId": "PRIV_CRED_123"
						}`,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "should return error when processing fails",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// Setup failing validation
				validation.On("ValidateCardInfoMessage", mock.AnythingOfType("*entities.PxpCardInfoMessage")).Return(errors.New("validation failed"))
			},
			event: events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId: "message-1",
						Body:      `{"invalid":"message"}`,
					},
				},
			},
			expectedError: true,
			errorContains: "failed to process SQS record 1",
		},
		{
			name: "should handle empty SQS event successfully",
			setupMocks: func(repo *MockCardInfoRepository, encryption *MockEncryptionService, validation *MockValidationService) {
				// No calls expected
			},
			event: events.SQSEvent{
				Records: []events.SQSMessage{},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockCardInfoRepository{}
			mockEncryption := &MockEncryptionService{}
			mockValidation := &MockValidationService{}
			mockLogger := mocks.GetMockLogger(t)

			tt.setupMocks(mockRepo, mockEncryption, mockValidation)

			// Create real use case with mocked dependencies
			useCase := use_cases.NewProcessCardInfoMessageUseCase(
				mockRepo,
				mockEncryption,
				mockValidation,
				mockLogger,
			)

			// Create adapter with real use case
			adapter := NewSQSAdapter(useCase, mockLogger)

			// Cast to concrete type to access HandleSQSEvent
			sqsAdapter := adapter.(*SQSAdapter)

			// Execute
			err := sqsAdapter.HandleSQSEvent(context.Background(), tt.event)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockEncryption.AssertExpectations(t)
			mockValidation.AssertExpectations(t)
		})
	}
}

func TestSQSAdapter_ImplementsMessageHandler(t *testing.T) {
	t.Run("should implement MessageHandler interface", func(t *testing.T) {
		// Setup with minimal mocks
		mockRepo := &MockCardInfoRepository{}
		mockEncryption := &MockEncryptionService{}
		mockValidation := &MockValidationService{}
		mockLogger := mocks.GetMockLogger(t)

		// Create real use case
		useCase := use_cases.NewProcessCardInfoMessageUseCase(
			mockRepo,
			mockEncryption,
			mockValidation,
			mockLogger,
		)

		// Create adapter
		adapter := NewSQSAdapter(useCase, mockLogger)

		// Assert - verify it implements the interface
		var _ interfaces.MessageHandler = adapter
	})
}

func TestNewSQSAdapter(t *testing.T) {
	t.Run("should create new SQS adapter", func(t *testing.T) {
		// Setup
		mockRepo := &MockCardInfoRepository{}
		mockEncryption := &MockEncryptionService{}
		mockValidation := &MockValidationService{}
		mockLogger := mocks.GetMockLogger(t)

		// Create real use case
		useCase := use_cases.NewProcessCardInfoMessageUseCase(
			mockRepo,
			mockEncryption,
			mockValidation,
			mockLogger,
		)

		// Execute
		adapter := NewSQSAdapter(useCase, mockLogger)

		// Assert
		assert.NotNil(t, adapter)

		// Verify it implements MessageHandler
		var _ interfaces.MessageHandler = adapter

		// Cast to concrete type to verify internal fields
		concreteAdapter := adapter.(*SQSAdapter)
		assert.NotNil(t, concreteAdapter.processCardInfoUseCase)
		assert.Equal(t, mockLogger, concreteAdapter.logger)
	})
}
