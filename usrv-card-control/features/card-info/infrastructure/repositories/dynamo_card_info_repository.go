package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/repositories"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo/builder"
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

const (
	EnvCardInfoTable = "DYNAMO_CARD_INFO_TABLE"

	ExternalReferenceIDField = "externalReferenceId"
	MerchantIDField          = "merchantId"
	ExpiresAtField           = "expiresAt"
)

// DynamoCardInfoRepository implements the CardInfoRepository using DynamoDB
type DynamoCardInfoRepository struct {
	dynamoGateway dynamo.IDynamoGateway
	logger        logger.KushkiLogger
	tableName     string
}

// NewDynamoCardInfoRepository creates a new DynamoDB repository instance
func NewDynamoCardInfoRepository(
	dynamoGateway dynamo.IDynamoGateway,
	logger logger.KushkiLogger,
) repositories.CardInfoRepository {
	return &DynamoCardInfoRepository{
		dynamoGateway: dynamoGateway,
		logger:        logger,
		tableName:     os.Getenv(EnvCardInfoTable),
	}
}

// Save stores the card information in DynamoDB
func (r *DynamoCardInfoRepository) Save(ctx context.Context, cardInfo *entities.StoredCardInfo) error {
	const operation = "DynamoCardInfoRepository.Save"

	r.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("ExternalReferenceID: %s", cardInfo.ExternalReferenceID))

	putBuilder := r.buildPutItemBuilder(cardInfo)

	if err := r.dynamoGateway.PutItem(ctx, putBuilder); err != nil {
		r.logger.Error(fmt.Sprintf("%s | Error", operation), err)
		return fmt.Errorf("failed to save card info to DynamoDB: %w", err)
	}

	r.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("ExternalReferenceID: %s", cardInfo.ExternalReferenceID))

	return nil
}

// FindByExternalReferenceID retrieves card information by external reference ID
func (r *DynamoCardInfoRepository) FindByExternalReferenceID(
	ctx context.Context,
	externalReferenceID string,
) (*entities.StoredCardInfo, error) {
	const operation = "DynamoCardInfoRepository.FindByExternalReferenceID"

	r.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("ExternalReferenceID: %s", externalReferenceID))

	getBuilder := r.buildGetItemBuilder(externalReferenceID)

	var cardInfo entities.StoredCardInfo
	if err := r.dynamoGateway.GetItem(ctx, getBuilder, &cardInfo); err != nil {
		if errors.Is(err, dynamoerror.ErrItemNotFound) {
			r.logger.Info(fmt.Sprintf("%s | NotFound", operation),
				fmt.Sprintf("ExternalReferenceID: %s", externalReferenceID))
			return nil, fmt.Errorf("card info not found for externalReferenceID: %s", externalReferenceID)
		}
		r.logger.Error(fmt.Sprintf("%s | Error", operation), err)
		return nil, fmt.Errorf("failed to get card info from DynamoDB: %w", err)
	}

	r.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("ExternalReferenceID: %s", externalReferenceID))

	return &cardInfo, nil
}

// FindByMerchantIDAndExternalReferenceID retrieves card information with merchant validation
func (r *DynamoCardInfoRepository) FindByMerchantIDAndExternalReferenceID(
	ctx context.Context,
	merchantID,
	externalReferenceID string,
) (*entities.StoredCardInfo, error) {
	const operation = "DynamoCardInfoRepository.FindByMerchantIDAndExternalReferenceID"

	r.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s, ExternalReferenceID: %s", merchantID, externalReferenceID))

	// Get the card info first
	cardInfo, err := r.FindByExternalReferenceID(ctx, externalReferenceID)
	if err != nil {
		return nil, err
	}

	// Validate merchant ownership
	if cardInfo.MerchantID != merchantID {
		r.logger.Info(fmt.Sprintf("%s | MerchantMismatch", operation),
			fmt.Sprintf("Expected: %s, Got: %s", merchantID, cardInfo.MerchantID))
		return nil, fmt.Errorf("card info not found for merchant: %s", merchantID)
	}

	return cardInfo, nil
}

// Delete removes card information from DynamoDB
func (r *DynamoCardInfoRepository) Delete(ctx context.Context, externalReferenceID string) error {
	const operation = "DynamoCardInfoRepository.Delete"

	r.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("ExternalReferenceID: %s", externalReferenceID))

	deleteBuilder := r.buildDeleteItemBuilder(externalReferenceID)

	if err := r.dynamoGateway.DeleteItem(ctx, deleteBuilder); err != nil {
		r.logger.Error(fmt.Sprintf("%s | Error", operation), err)
		return fmt.Errorf("failed to delete card info from DynamoDB: %w", err)
	}

	r.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("ExternalReferenceID: %s", externalReferenceID))

	return nil
}

// FindExpiredRecords finds records that have exceeded the 180-day limit
func (r *DynamoCardInfoRepository) FindExpiredRecords(
	ctx context.Context,
	currentTime int64,
) ([]*entities.StoredCardInfo, error) {
	const operation = "DynamoCardInfoRepository.FindExpiredRecords"

	r.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("CurrentTime: %d", currentTime))

	scanBuilder := r.buildScanExpiredBuilder(currentTime)

	var expiredRecords []*entities.StoredCardInfo
	if err := r.dynamoGateway.ScanItems(ctx, scanBuilder, &expiredRecords); err != nil {
		r.logger.Error(fmt.Sprintf("%s | Error", operation), err)
		return nil, fmt.Errorf("failed to scan expired records: %w", err)
	}

	r.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("Found %d expired records", len(expiredRecords)))

	return expiredRecords, nil
}

// Builder methods following the existing project patterns

// buildPutItemBuilder creates a put item builder for saving card info
func (r *DynamoCardInfoRepository) buildPutItemBuilder(cardInfo *entities.StoredCardInfo) *builder.PutItemBuilder {
	return builder.NewPutItemBuilder().
		WithItem(cardInfo).
		WithTable(r.tableName)
}

// buildGetItemBuilder creates a get item builder for retrieving card info
func (r *DynamoCardInfoRepository) buildGetItemBuilder(externalReferenceID string) *builder.GetItemBuilder {
	return builder.NewGetItemBuilder().
		WithTable(r.tableName).
		WithPartitionKey(ExternalReferenceIDField, externalReferenceID).
		WithConsistentRead(true)
}

// buildDeleteItemBuilder creates a delete item builder for removing card info
func (r *DynamoCardInfoRepository) buildDeleteItemBuilder(externalReferenceID string) *builder.DeleteItemBuilder {
	return builder.NewDeleteItemBuilder().
		WithTable(r.tableName).
		WithPartitionKey(ExternalReferenceIDField, externalReferenceID)
}

// buildScanExpiredBuilder creates a scan builder for finding expired records
func (r *DynamoCardInfoRepository) buildScanExpiredBuilder(currentTime int64) *builder.ScanBuilder {
	filter := expression.Name(ExpiresAtField).LessThan(expression.Value(currentTime))

	expr := expression.NewBuilder().WithFilter(filter)

	return builder.NewScanBuilder().
		WithTable(r.tableName).
		WithExpression(&expr)
}
