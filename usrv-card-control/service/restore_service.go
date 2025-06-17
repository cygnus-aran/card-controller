package service

import (
	"context"
	"fmt"

	"bitbucket.org/kushki/usrv-card-control/gateway"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/aws/aws-lambda-go/events"
)

type IRestoreService interface {
	RestoreDailyRetries(ctx context.Context, event events.SQSEvent) error
}

// RestoreService to clean retries.
type RestoreService struct {
	Logger logger.KushkiLogger
	Dynamo dynamo.IDynamoGateway
}

const restoreSrvTag = "RestoreService | %s"

// RefNewRestoreService ref to new service.
var RefNewRestoreService = NewRestoreService

// NewRestoreService function to instantiate.
func NewRestoreService(kskLogger logger.KushkiLogger, dynamoGtw dynamo.IDynamoGateway) IRestoreService {
	return &RestoreService{
		Logger: kskLogger,
		Dynamo: dynamoGtw,
	}
}

// InitRestoreService used to initialize dependencies for service.
func InitRestoreService(ctx context.Context, event events.SQSEvent) error {
	kskLogger := newKushkiLogger(ctx)
	dynamoGtw, err := initializeDynamoGtw(ctx, kskLogger)
	if err != nil {
		kskLogger.Error(fmt.Sprintf(blockSrvTag, "Error intializing dynamo"), err)
		return err
	}

	service := RefNewRestoreService(kskLogger, dynamoGtw)

	return service.RestoreDailyRetries(ctx, event)
}

func (rs *RestoreService) RestoreDailyRetries(ctx context.Context, event events.SQSEvent) error {
	var request types.RestoreDailyRequest
	if err := jsonUnmarshalCaller([]byte(event.Records[0].Body), &request); err != nil {
		return err
	}

	if err := rs.cleanDailyRetries(ctx, request); err != nil {
		return err
	}

	return rs.cleanLastRetry(ctx, request)
}

func (rs *RestoreService) cleanDailyRetries(ctx context.Context, request types.RestoreDailyRequest) error {
	rs.Logger.Info(fmt.Sprintf(restoreSrvTag, "cleanDailyRetries"), "[CLEANING]")
	queryInput := gateway.QueryRetriesBuilder(request.CardID, request.MerchantID)
	var out []types.CardRetry
	if err := rs.Dynamo.Query(ctx, queryInput, &out); err != nil {
		return err
	}

	for _, retry := range out {
		deleteItemBuilder := gateway.DeleteCardRetryBuilder(retry.RetryKey)
		if err := rs.Dynamo.DeleteItem(ctx, deleteItemBuilder); err != nil {
			return err
		}
	}

	return nil
}

func (rs *RestoreService) cleanLastRetry(ctx context.Context, request types.RestoreDailyRequest) error {
	rs.Logger.Info(fmt.Sprintf(restoreSrvTag, "cleanLastRetry"), "[CLEANING]")
	input := gateway.GetBlockedCardBuilder(request.CardID)
	var blockedCard types.DynamoBlockedCard

	if err := rs.Dynamo.GetItem(ctx, input, &blockedCard); err != nil {
		return err
	}

	cleanInput := gateway.UpdateLastRetryBuilder(0, request.MerchantID, blockedCard)
	return rs.Dynamo.UpdateItem(ctx, cleanInput)
}
