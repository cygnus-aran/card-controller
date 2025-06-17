package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	constants "bitbucket.org/kushki/usrv-card-control"
	dynamoBuilders "bitbucket.org/kushki/usrv-card-control/gateway"
	"bitbucket.org/kushki/usrv-card-control/types"
	errorsCore "bitbucket.org/kushki/usrv-go-core/errors"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-lambda-go/events"
)

const (
	checkCardStatusServiceTag = "CheckCardStatusService"
)

var (
	refNewCheckCardStatusService = NewCheckCardStatusService
)

type CheckCardStatusService struct {
	Context context.Context
	Logger  logger.KushkiLogger
	Dynamo  dynamo.IDynamoGateway
}

// ICheckCardStatusService Interface to get SyncMerchant.
type ICheckCardStatusService interface {
	CheckCardStatus(checkCardStatusRequest types.CheckCardStatusRequest) types.CheckCardStatusResponse
}

// NewCheckCardStatusService service initiator.
// Should initialize the dependencies for this service.
func NewCheckCardStatusService(ctx context.Context, dynamo dynamo.IDynamoGateway, lg logger.KushkiLogger) ICheckCardStatusService {
	return &CheckCardStatusService{
		Context: ctx,
		Dynamo:  dynamo,
		Logger:  lg,
	}
}

// InitializeCheckCardStatus init process sync status merchant in elastic.
func InitializeCheckCardStatus(ctx context.Context, event events.APIGatewayProxyRequest) (types.CheckCardStatusResponse, error) {
	tag := fmt.Sprint(checkCardStatusServiceTag, "InitializeCheckCardStatus")

	kskLogger := newKushkiLogger(ctx)

	var checkCardStatusRequest types.CheckCardStatusRequest
	if err := jsonUnmarshalCaller([]byte(event.Body), &checkCardStatusRequest); err != nil {
		kskLogger.Error(fmt.Sprintf(tag, "Error while unmarshal check card status request: "), err)
		return types.CheckCardStatusResponse{}, errorsCore.NewKushkiError(errorsCore.Errors[errorsCore.E001], gabs.Wrap(errorsCore.GenericErrorMetadata{
			Message: err.Error(), Origin: tag}))
	}
	dynamoGtw, err := initializeDynamoGtw(ctx, kskLogger)
	if err != nil {
		kskLogger.Error(fmt.Sprintf(tag, "Error initializing dynamo: "), err)
		return types.CheckCardStatusResponse{}, errorsCore.NewKushkiError(errorsCore.Errors[errorsCore.E002], gabs.Wrap(errorsCore.GenericErrorMetadata{
			Message: err.Error(), Origin: tag}))
	}

	checkCardStatusService := refNewCheckCardStatusService(ctx, dynamoGtw, kskLogger)

	return checkCardStatusService.CheckCardStatus(checkCardStatusRequest), nil
}

func (s *CheckCardStatusService) CheckCardStatus(checkCardStatusRequest types.CheckCardStatusRequest) types.CheckCardStatusResponse {
	tag := fmt.Sprintf("%s | %s", checkCardStatusServiceTag, "checkCardStatus")
	if checkCardStatusRequest.CardID == "" {
		s.Logger.Info(fmt.Sprintf(tag, "omit logic"), "EMPTY CARD_ID")
		return types.CheckCardStatusResponse{}
	}

	blockedCardInfo, err := s.getBlockedCardInfo(checkCardStatusRequest.CardID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf(tag, "error getting blocked card info: "), err)
		return types.CheckCardStatusResponse{}
	}

	lockInfo, isBlocked := blockedCardInfo.BlockedMerchants[checkCardStatusRequest.MerchantIdentifier]
	if !isBlocked {
		return types.CheckCardStatusResponse{}
	}

	if strings.EqualFold(lockInfo.BlockType, constants.PERMANENT) {
		return types.CheckCardStatusResponse{Blocked: true, BlockType: constants.PERMANENT, HasRetries: false}
	}

	currentDate := time.Now().UnixMilli()
	isTemporarilyBlocked := lockInfo.ExpirationDate > currentDate
	lastRetry := time.UnixMilli(lockInfo.LastRetry).Add(24 * time.Hour).UnixMilli()
	hasRetries := lastRetry > currentDate
	return types.CheckCardStatusResponse{
		Blocked:    isTemporarilyBlocked,
		BlockType:  getBlockType(isTemporarilyBlocked),
		HasRetries: hasRetries,
	}
}

// getBlockedCardInfo get block card info from dynamodb.
func (s *CheckCardStatusService) getBlockedCardInfo(cardID string) (types.DynamoBlockedCard, error) {
	builder := dynamoBuilders.GetBlockedCardBuilder(cardID)
	blockedCardInfo := types.DynamoBlockedCard{}
	if err := s.Dynamo.GetItem(s.Context, builder, &blockedCardInfo); err != nil {
		return blockedCardInfo, err
	}
	return blockedCardInfo, nil
}

func getBlockType(isTemporarily bool) string {
	if isTemporarily {
		return constants.TEMPORARY
	}
	return ""
}
