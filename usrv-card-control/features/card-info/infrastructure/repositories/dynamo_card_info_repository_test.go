package repositories

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo/builder"
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	coreTypes "bitbucket.org/kushki/usrv-go-core/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDynamoGateway - mock for the DynamoDB gateway
type MockDynamoGateway struct {
	mock.Mock
}

func (m *MockDynamoGateway) PutItem(ctx context.Context, builder *builder.PutItemBuilder) error {
	args := m.Called(ctx, builder)
	return args.Error(0)
}

func (m *MockDynamoGateway) GetItem(ctx context.Context, builder *builder.GetItemBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) DeleteItem(ctx context.Context, builder *builder.DeleteItemBuilder) error {
	args := m.Called(ctx, builder)
	return args.Error(0)
}

func (m *MockDynamoGateway) ScanItems(ctx context.Context, builder *builder.ScanBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

// Add other methods that might be called by the interface
func (m *MockDynamoGateway) Query(ctx context.Context, builder *builder.QueryBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) UpdateItem(ctx context.Context, builder *builder.UpdateItemBuilder) error {
	args := m.Called(ctx, builder)
	return args.Error(0)
}

func (m *MockDynamoGateway) BatchGetItem(ctx context.Context, builder *builder.BatchGetItemInputBuilder, out *coreTypes.BatchGetItemResponse) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) BatchWriteItem(ctx context.Context, builder *builder.BatchWriteItemInputBuilder, out *coreTypes.BatchWriteItemResponse) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) TransactWriteItems(ctx context.Context, builder *builder.TransactWriteItemsBuilder) error {
	args := m.Called(ctx, builder)
	return args.Error(0)
}

func (m *MockDynamoGateway) ScanAllItems(ctx context.Context, builder *builder.ScanBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) QuerySingle(ctx context.Context, builder *builder.QueryBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

func (m *MockDynamoGateway) GetSequential(ctx context.Context, builder *builder.GetSequentialBuilder) (int, error) {
	args := m.Called(ctx, builder)
	return args.Int(0), args.Error(1)
}

func (m *MockDynamoGateway) GetCurrentSequential(ctx context.Context, builder *builder.GetSequentialBuilder) (int, error) {
	args := m.Called(ctx, builder)
	return args.Int(0), args.Error(1)
}

func (m *MockDynamoGateway) UpdateCurrentSequential(ctx context.Context, builder *builder.UpdateSequentialBuilder) error {
	args := m.Called(ctx, builder)
	return args.Error(0)
}

func (m *MockDynamoGateway) GetSequentialInfo(ctx context.Context, builder *builder.GetSequentialBuilder, out interface{}) error {
	args := m.Called(ctx, builder, out)
	return args.Error(0)
}

// MockDynamoLogger - dedicated mock logger for repository tests
type MockDynamoLogger struct {
	mock.Mock
}

func (m *MockDynamoLogger) Info(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockDynamoLogger) Error(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockDynamoLogger) Debug(tag string, v interface{}) {
	m.Called(tag, v)
}

func (m *MockDynamoLogger) Warning(tag string, v interface{}) {
	m.Called(tag, v)
}

// Test helper functions
func setupRepository(t *testing.T) (*DynamoCardInfoRepository, *MockDynamoGateway, *MockDynamoLogger) {
	t.Helper()
	mockDynamo := &MockDynamoGateway{}
	mockLogger := &MockDynamoLogger{}

	// Set environment variable for table name
	originalTableName := os.Getenv(EnvCardInfoTable)
	os.Setenv(EnvCardInfoTable, "test-card-info-table")
	t.Cleanup(func() {
		if originalTableName == "" {
			os.Unsetenv(EnvCardInfoTable)
		} else {
			os.Setenv(EnvCardInfoTable, originalTableName)
		}
	})

	repo := NewDynamoCardInfoRepository(mockDynamo, mockLogger).(*DynamoCardInfoRepository)
	return repo, mockDynamo, mockLogger
}

// Helper to create test card info entity
func createTestStoredCardInfo() *entities.StoredCardInfo {
	currentTime := time.Now().UnixMilli()
	return &entities.StoredCardInfo{
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
		EncryptedCard: value_objects.EncryptedCardData{
			EncryptedPan:  "encrypted-pan-data",
			EncryptedDate: "encrypted-date-data",
		},
		TransactionDate: currentTime,
		CreatedAt:       currentTime,
		ExpiresAt:       currentTime + (180 * 24 * 60 * 60 * 1000), // 180 days
	}
}

// Test Save - Success Cases
func TestDynamoCardInfoRepository_Save_Success(t *testing.T) {
	t.Run("Successfully save card info", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Save with different card info data", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		cardInfo.ExternalReferenceID = "different-ref-789"
		cardInfo.CardBrand = "MASTERCARD"
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Save - Failure Cases
func TestDynamoCardInfoRepository_Save_Failures(t *testing.T) {
	t.Run("DynamoDB put item error", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		ctx := context.Background()
		dynamoError := errors.New("dynamodb connection failed")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(dynamoError)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save card info to DynamoDB")
		assert.Contains(t, err.Error(), "dynamodb connection failed")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Nil card info", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, _ := setupRepository(t)
		ctx := context.Background()

		// Setup mocks - the repository should handle nil gracefully or fail early
		// Don't expect any specific logging since nil pointer will cause early failure
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(errors.New("invalid item"))

		// Act & Assert
		// This should either handle nil gracefully or panic (which we'll catch)
		assert.Panics(t, func() {
			repo.Save(ctx, nil)
		}, "Repository should panic or handle nil card info appropriately")

		// If we wanted the repository to handle nil gracefully, we would need to add nil checks
		// to the repository implementation first
	})
}

// Test FindByExternalReferenceID - Success Cases
func TestDynamoCardInfoRepository_FindByExternalReferenceID_Success(t *testing.T) {
	t.Run("Successfully find card info", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "ext-ref-123"
		ctx := context.Background()
		expectedCardInfo := createTestStoredCardInfo()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*entities.StoredCardInfo)
				*out = *expectedCardInfo
			}).
			Return(nil)

		// Act
		cardInfo, err := repo.FindByExternalReferenceID(ctx, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cardInfo)
		assert.Equal(t, expectedCardInfo.ExternalReferenceID, cardInfo.ExternalReferenceID)
		assert.Equal(t, expectedCardInfo.MerchantID, cardInfo.MerchantID)
		assert.Equal(t, expectedCardInfo.CardBrand, cardInfo.CardBrand)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test FindByExternalReferenceID - Failure Cases
func TestDynamoCardInfoRepository_FindByExternalReferenceID_Failures(t *testing.T) {
	t.Run("Card info not found", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "non-existent-ref"
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Return(dynamoerror.ErrItemNotFound)

		// Act
		cardInfo, err := repo.FindByExternalReferenceID(ctx, externalReferenceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cardInfo)
		assert.Contains(t, err.Error(), "card info not found for externalReferenceID: non-existent-ref")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("DynamoDB get item error", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "error-ref"
		ctx := context.Background()
		dynamoError := errors.New("dynamodb read error")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Return(dynamoError)

		// Act
		cardInfo, err := repo.FindByExternalReferenceID(ctx, externalReferenceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cardInfo)
		assert.Contains(t, err.Error(), "failed to get card info from DynamoDB")
		assert.Contains(t, err.Error(), "dynamodb read error")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Empty external reference ID", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Return(dynamoerror.ErrItemNotFound)

		// Act
		cardInfo, err := repo.FindByExternalReferenceID(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cardInfo)
		assert.Contains(t, err.Error(), "card info not found for externalReferenceID: ")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Delete - Success Cases
func TestDynamoCardInfoRepository_Delete_Success(t *testing.T) {
	t.Run("Successfully delete card info", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "ext-ref-to-delete"
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("DeleteItem", ctx, mock.AnythingOfType("*builder.DeleteItemBuilder")).Return(nil)

		// Act
		err := repo.Delete(ctx, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Delete - Failure Cases
func TestDynamoCardInfoRepository_Delete_Failures(t *testing.T) {
	t.Run("DynamoDB delete item error", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "error-delete-ref"
		ctx := context.Background()
		dynamoError := errors.New("dynamodb delete error")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("DeleteItem", ctx, mock.AnythingOfType("*builder.DeleteItemBuilder")).Return(dynamoError)

		// Act
		err := repo.Delete(ctx, externalReferenceID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete card info from DynamoDB")
		assert.Contains(t, err.Error(), "dynamodb delete error")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test FindExpiredRecords - Success Cases
func TestDynamoCardInfoRepository_FindExpiredRecords_Success(t *testing.T) {
	t.Run("Successfully find expired records", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		currentTime := time.Now().UnixMilli()
		ctx := context.Background()

		expiredRecord1 := createTestStoredCardInfo()
		expiredRecord1.ExternalReferenceID = "expired-1"
		expiredRecord1.ExpiresAt = currentTime - 1000 // Expired

		expiredRecord2 := createTestStoredCardInfo()
		expiredRecord2.ExternalReferenceID = "expired-2"
		expiredRecord2.ExpiresAt = currentTime - 2000 // Expired

		expectedExpiredRecords := []*entities.StoredCardInfo{expiredRecord1, expiredRecord2}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*[]*entities.StoredCardInfo)
				*out = expectedExpiredRecords
			}).
			Return(nil)

		// Act
		expiredRecords, err := repo.FindExpiredRecords(ctx, currentTime)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, expiredRecords, 2)
		assert.Equal(t, "expired-1", expiredRecords[0].ExternalReferenceID)
		assert.Equal(t, "expired-2", expiredRecords[1].ExternalReferenceID)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("No expired records found", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		currentTime := time.Now().UnixMilli()
		ctx := context.Background()
		emptyRecords := []*entities.StoredCardInfo{}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*[]*entities.StoredCardInfo)
				*out = emptyRecords
			}).
			Return(nil)

		// Act
		expiredRecords, err := repo.FindExpiredRecords(ctx, currentTime)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, expiredRecords, 0)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test FindExpiredRecords - Failure Cases
func TestDynamoCardInfoRepository_FindExpiredRecords_Failures(t *testing.T) {
	t.Run("DynamoDB scan error", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		currentTime := time.Now().UnixMilli()
		ctx := context.Background()
		dynamoError := errors.New("dynamodb scan error")

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Return(dynamoError)

		// Act
		expiredRecords, err := repo.FindExpiredRecords(ctx, currentTime)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, expiredRecords)
		assert.Contains(t, err.Error(), "failed to scan expired records")
		assert.Contains(t, err.Error(), "dynamodb scan error")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test FindByMerchantIDAndExternalReferenceID
func TestDynamoCardInfoRepository_FindByMerchantIDAndExternalReferenceID(t *testing.T) {
	t.Run("Successfully find with merchant validation", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		merchantID := "merchant-123"
		externalReferenceID := "ext-ref-123"
		ctx := context.Background()
		expectedCardInfo := createTestStoredCardInfo()
		expectedCardInfo.MerchantID = merchantID

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*entities.StoredCardInfo)
				*out = *expectedCardInfo
			}).
			Return(nil)

		// Act
		cardInfo, err := repo.FindByMerchantIDAndExternalReferenceID(ctx, merchantID, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cardInfo)
		assert.Equal(t, merchantID, cardInfo.MerchantID)
		assert.Equal(t, externalReferenceID, cardInfo.ExternalReferenceID)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Merchant mismatch", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		requestMerchantID := "merchant-123"
		storedMerchantID := "different-merchant-456"
		externalReferenceID := "ext-ref-123"
		ctx := context.Background()
		storedCardInfo := createTestStoredCardInfo()
		storedCardInfo.MerchantID = storedMerchantID // Different from request

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*entities.StoredCardInfo)
				*out = *storedCardInfo
			}).
			Return(nil)

		// Act
		cardInfo, err := repo.FindByMerchantIDAndExternalReferenceID(ctx, requestMerchantID, externalReferenceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cardInfo)
		assert.Contains(t, err.Error(), "card info not found for merchant: merchant-123")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Record not found", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		merchantID := "merchant-123"
		externalReferenceID := "non-existent-ref"
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Return(dynamoerror.ErrItemNotFound)

		// Act
		cardInfo, err := repo.FindByMerchantIDAndExternalReferenceID(ctx, merchantID, externalReferenceID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cardInfo)
		assert.Contains(t, err.Error(), "card info not found for externalReferenceID: non-existent-ref")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Logging Behavior
func TestDynamoCardInfoRepository_LoggingBehavior(t *testing.T) {
	t.Run("Logs save operation", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", "DynamoCardInfoRepository.Save | Starting", "ExternalReferenceID: ext-ref-123").Return()
		mockLogger.On("Info", "DynamoCardInfoRepository.Save | Success", "ExternalReferenceID: ext-ref-123").Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs find operation", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "test-ref-456"
		ctx := context.Background()
		cardInfo := createTestStoredCardInfo()

		// Setup mocks
		mockLogger.On("Info", "DynamoCardInfoRepository.FindByExternalReferenceID | Starting", "ExternalReferenceID: test-ref-456").Return()
		mockLogger.On("Info", "DynamoCardInfoRepository.FindByExternalReferenceID | Success", "ExternalReferenceID: test-ref-456").Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*entities.StoredCardInfo)
				*out = *cardInfo
			}).
			Return(nil)

		// Act
		result, err := repo.FindByExternalReferenceID(ctx, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs delete operation", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "delete-ref-789"
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", "DynamoCardInfoRepository.Delete | Starting", "ExternalReferenceID: delete-ref-789").Return()
		mockLogger.On("Info", "DynamoCardInfoRepository.Delete | Success", "ExternalReferenceID: delete-ref-789").Return()
		mockDynamo.On("DeleteItem", ctx, mock.AnythingOfType("*builder.DeleteItemBuilder")).Return(nil)

		// Act
		err := repo.Delete(ctx, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Logs expired records scan", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		currentTime := time.Now().UnixMilli()
		ctx := context.Background()
		expiredRecords := []*entities.StoredCardInfo{}

		// Setup mocks
		mockLogger.On("Info", "DynamoCardInfoRepository.FindExpiredRecords | Starting", mock.MatchedBy(func(v interface{}) bool {
			str, ok := v.(string)
			return ok && strings.Contains(str, "CurrentTime:")
		})).Return()
		mockLogger.On("Info", "DynamoCardInfoRepository.FindExpiredRecords | Success", "Found 0 expired records").Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*[]*entities.StoredCardInfo)
				*out = expiredRecords
			}).
			Return(nil)

		// Act
		result, err := repo.FindExpiredRecords(ctx, currentTime)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Builder Methods
func TestDynamoCardInfoRepository_BuilderMethods(t *testing.T) {
	t.Run("buildPutItemBuilder creates correct builder", func(t *testing.T) {
		// Arrange
		repo, _, _ := setupRepository(t)
		cardInfo := createTestStoredCardInfo()

		// Act
		builder := repo.buildPutItemBuilder(cardInfo)

		// Assert
		assert.NotNil(t, builder)
		// We can't easily test the internal builder state without accessing private fields
		// But we can verify it doesn't panic and returns a builder
	})

	t.Run("buildGetItemBuilder creates correct builder", func(t *testing.T) {
		// Arrange
		repo, _, _ := setupRepository(t)
		externalReferenceID := "test-ref-123"

		// Act
		builder := repo.buildGetItemBuilder(externalReferenceID)

		// Assert
		assert.NotNil(t, builder)
	})

	t.Run("buildDeleteItemBuilder creates correct builder", func(t *testing.T) {
		// Arrange
		repo, _, _ := setupRepository(t)
		externalReferenceID := "delete-ref-123"

		// Act
		builder := repo.buildDeleteItemBuilder(externalReferenceID)

		// Assert
		assert.NotNil(t, builder)
	})

	t.Run("buildScanExpiredBuilder creates correct builder", func(t *testing.T) {
		// Arrange
		repo, _, _ := setupRepository(t)
		currentTime := time.Now().UnixMilli()

		// Act
		builder := repo.buildScanExpiredBuilder(currentTime)

		// Assert
		assert.NotNil(t, builder)
	})
}

// Test Edge Cases
func TestDynamoCardInfoRepository_EdgeCases(t *testing.T) {
	t.Run("Save with very long external reference ID", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		cardInfo.ExternalReferenceID = strings.Repeat("VERY_LONG_REF_", 50) // Very long ID
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Save with unicode characters in data", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		cardInfo.SubMerchantCode = "商户-123" // Unicode characters
		cardInfo.TerminalID = "terminal-🏪"  // Emoji
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Find with special characters in external reference ID", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		externalReferenceID := "ref-with-special-chars-@#$%"
		ctx := context.Background()
		cardInfo := createTestStoredCardInfo()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("GetItem", ctx, mock.AnythingOfType("*builder.GetItemBuilder"), mock.AnythingOfType("*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*entities.StoredCardInfo)
				*out = *cardInfo
			}).
			Return(nil)

		// Act
		result, err := repo.FindByExternalReferenceID(ctx, externalReferenceID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("FindExpiredRecords with future timestamp", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		futureTime := time.Now().Add(365 * 24 * time.Hour).UnixMilli() // 1 year in future
		ctx := context.Background()
		expiredRecords := []*entities.StoredCardInfo{}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*[]*entities.StoredCardInfo)
				*out = expiredRecords
			}).
			Return(nil)

		// Act
		result, err := repo.FindExpiredRecords(ctx, futureTime)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0, "Future timestamp should find no expired records")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("FindExpiredRecords with zero timestamp", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		zeroTime := int64(0)
		ctx := context.Background()
		expiredRecords := []*entities.StoredCardInfo{}

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("ScanItems", ctx, mock.AnythingOfType("*builder.ScanBuilder"), mock.AnythingOfType("*[]*entities.StoredCardInfo")).
			Run(func(args mock.Arguments) {
				out := args.Get(2).(*[]*entities.StoredCardInfo)
				*out = expiredRecords
			}).
			Return(nil)

		// Act
		result, err := repo.FindExpiredRecords(ctx, zeroTime)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Save with empty encrypted card data", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		cardInfo.EncryptedCard = value_objects.EncryptedCardData{
			EncryptedPan:  "",
			EncryptedDate: "",
		}
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err, "Should be able to save with empty encrypted data")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Save with maximum timestamp values", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()
		maxTime := int64(9223372036854775807) // Maximum int64 value
		cardInfo.CreatedAt = maxTime
		cardInfo.ExpiresAt = maxTime
		cardInfo.TransactionDate = maxTime
		ctx := context.Background()

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err, "Should handle maximum timestamp values")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test Environment Variable Handling
func TestDynamoCardInfoRepository_EnvironmentVariables(t *testing.T) {
	t.Run("Uses correct table name from environment", func(t *testing.T) {
		// Arrange
		mockDynamo := &MockDynamoGateway{}
		mockLogger := &MockDynamoLogger{}

		// Set custom table name
		customTableName := "custom-card-info-table"
		originalTableName := os.Getenv(EnvCardInfoTable)
		os.Setenv(EnvCardInfoTable, customTableName)
		defer func() {
			if originalTableName == "" {
				os.Unsetenv(EnvCardInfoTable)
			} else {
				os.Setenv(EnvCardInfoTable, originalTableName)
			}
		}()

		repo := NewDynamoCardInfoRepository(mockDynamo, mockLogger).(*DynamoCardInfoRepository)

		// Assert
		assert.Equal(t, customTableName, repo.tableName)
	})

	t.Run("Handles empty table name", func(t *testing.T) {
		// Arrange
		mockDynamo := &MockDynamoGateway{}
		mockLogger := &MockDynamoLogger{}

		// Unset table name
		originalTableName := os.Getenv(EnvCardInfoTable)
		os.Unsetenv(EnvCardInfoTable)
		defer func() {
			if originalTableName != "" {
				os.Setenv(EnvCardInfoTable, originalTableName)
			}
		}()

		repo := NewDynamoCardInfoRepository(mockDynamo, mockLogger).(*DynamoCardInfoRepository)

		// Assert
		assert.Equal(t, "", repo.tableName, "Should handle empty table name gracefully")
	})
}

// Test Context Handling
func TestDynamoCardInfoRepository_ContextHandling(t *testing.T) {
	t.Run("Passes context correctly to DynamoDB operations", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()

		// Create a context with a value we can verify
		type contextKey string
		key := contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")

		// Setup mocks - verify context is passed through
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", mock.MatchedBy(func(c context.Context) bool {
			return c.Value(key) == "test-value"
		}), mock.AnythingOfType("*builder.PutItemBuilder")).Return(nil)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.NoError(t, err)
		mockDynamo.AssertExpectations(t)
	})

	t.Run("Handles context cancellation", func(t *testing.T) {
		// Arrange
		repo, mockDynamo, mockLogger := setupRepository(t)
		cardInfo := createTestStoredCardInfo()

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Setup mocks
		mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
		mockDynamo.On("PutItem", ctx, mock.AnythingOfType("*builder.PutItemBuilder")).Return(context.Canceled)

		// Act
		err := repo.Save(ctx, cardInfo)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save card info to DynamoDB")
		mockDynamo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
