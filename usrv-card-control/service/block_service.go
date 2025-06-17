package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/gateway"
	"bitbucket.org/kushki/usrv-card-control/types"
	core "bitbucket.org/kushki/usrv-go-core"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/aws/aws-lambda-go/events"
)

type IBlockService interface {
	ProcessBlock(ctx context.Context, event events.SQSEvent) error
}

// BlockService to manage card blocks and retries.
type BlockService struct {
	Logger logger.KushkiLogger
	Dynamo dynamo.IDynamoGateway
}

const blockSrvTag = "BlockService | %s"

// RefNewBlockService ref to new service.
var RefNewBlockService = NewBlockService

// NewBlockService function to instantiate.
func NewBlockService(kskLogger logger.KushkiLogger, dynamoGtw dynamo.IDynamoGateway) IBlockService {
	return &BlockService{
		Logger: kskLogger,
		Dynamo: dynamoGtw,
	}
}

// InitBlockService used to initialize dependencies for service.
func InitBlockService(ctx context.Context, event events.SQSEvent) error {
	kskLogger := newKushkiLogger(ctx)
	dynamoGtw, err := initializeDynamoGtw(ctx, kskLogger)
	if err != nil {
		kskLogger.Error(fmt.Sprintf(blockSrvTag, "Error intializing dynamo"), err)
		return err
	}

	service := RefNewBlockService(kskLogger, dynamoGtw)

	return service.ProcessBlock(ctx, event)
}

// ProcessBlock block or increment card retries.
func (bs *BlockService) ProcessBlock(ctx context.Context, event events.SQSEvent) error {
	var request types.BlockCardRequest
	err := jsonUnmarshalCaller([]byte(event.Records[0].Body), &request)
	if err != nil {
		return err
	}

	if request.CardID == "" {
		bs.info("empty required request params", "[SKIPPING LOGIC]")
		return nil
	}

	blockedCard, err := bs.getBlockedCard(ctx, request.CardID)
	if errors.Is(err, dynamoerror.ErrItemNotFound) {
		bs.info("No block found", "[NEW Block]")
		blockedCard, err = bs.generateNewBlockedCard(ctx, request)
	}
	if err != nil {
		return err
	}

	if strings.EqualFold(request.Operation, constants.BlockCardOperation) {
		return bs.blockCard(ctx, request, blockedCard)
	}

	currentDate := time.Now().UTC().UnixMilli()
	bs.info("ProcessBlock", "[CHECKING RETRIES...]")
	frequencies := constants.Frequencies[request.Franchise]
	blockedMap := make(map[string]bool)
	for _, frequency := range frequencies {
		blockedMap[frequency], err = bs.processRetry(ctx, request, frequency, currentDate)
		if err != nil {
			return err
		}
	}

	blocked := getBlocked(blockedMap)
	if blocked {
		bs.info("Process Retry - Retries exceeded limit", "[BLOCKING CARD]")
		return bs.blockCard(ctx, request, blockedCard)
	}

	// VISA does not have daily retries.
	if strings.EqualFold(request.Franchise, core.BrandVisa) {
		return nil
	}
	return bs.updateLastRetry(ctx, request.MerchantIdentifier, currentDate, blockedCard)
}

func getBlocked(blockedMap map[string]bool) bool {
	for _, blocked := range blockedMap {
		if blocked {
			return true
		}
	}
	return false
}

func (bs *BlockService) generateNewBlockedCard(ctx context.Context, request types.BlockCardRequest) (types.DynamoBlockedCard, error) {
	blockedCard := types.DynamoBlockedCard{
		BlockedMerchants: make(map[string]types.BlockedMerchant),
		CardID:           request.CardID,
		TimeStamp:        time.Now().UTC().UnixMilli(),
	}
	input := gateway.PutBlockedCardBuilder(blockedCard)

	return blockedCard, bs.Dynamo.PutItem(ctx, input)
}

func (bs *BlockService) processRetry(ctx context.Context, request types.BlockCardRequest, frequency string, currentDate int64) (blocked bool, err error) {
	key := generateRetryKey(request, frequency)

	retry, err := bs.getRetry(ctx, key)
	if err != nil && !errors.Is(err, dynamoerror.ErrItemNotFound) {
		return false, err
	}

	bs.info("Process Retry", "[Incrementing current retry]")
	validRetries, err := bs.incrementRetry(ctx, request, frequency, retry, currentDate)
	if err != nil {
		return false, err
	}

	limit := constants.Limits[request.Franchise][frequency]
	if len(validRetries) >= limit {
		blocked = true
	}

	return blocked, nil
}

func (bs *BlockService) updateLastRetry(
	ctx context.Context,
	merchantID string,
	currentDate int64,
	blockedCard types.DynamoBlockedCard,
) error {
	bs.info("updateLastRetry", "[updating]")
	input := gateway.UpdateLastRetryBuilder(currentDate, merchantID, blockedCard)

	return bs.Dynamo.UpdateItem(ctx, input)
}

func (bs *BlockService) incrementRetry(
	ctx context.Context,
	request types.BlockCardRequest,
	frequency string,
	cardRetry types.CardRetry,
	currentDate int64) ([]int64, error) {
	retries := getValidRetries(currentDate, cardRetry.Retries, frequency)
	bs.info("incrementRetry | VALID RETRIES", retries)
	key := generateRetryKey(request, frequency)
	input := gateway.IncrementRetryBuilder(retries, request, key, cardRetry)

	err := bs.Dynamo.UpdateItem(ctx, input)

	return retries, err
}

func (bs *BlockService) getRetry(ctx context.Context, key string) (types.CardRetry, error) {
	getItem := gateway.GetRetryBuilder(key)

	var out types.CardRetry

	err := bs.Dynamo.GetItem(ctx, getItem, &out)

	return out, err
}

func (bs *BlockService) getBlockedCard(ctx context.Context, cardID string) (types.DynamoBlockedCard, error) {
	getItem := gateway.GetBlockedCardBuilder(cardID)

	var out types.DynamoBlockedCard

	err := bs.Dynamo.GetItem(ctx, getItem, &out)

	return out, err
}

func (bs *BlockService) blockCard(
	ctx context.Context,
	request types.BlockCardRequest,
	blockedCard types.DynamoBlockedCard) error {
	item := gateway.UpdateBlockCardBuilder(request, blockedCard)

	return bs.Dynamo.UpdateItem(ctx, item)
}

func (bs *BlockService) info(process string, v interface{}) {
	bs.Logger.Info(fmt.Sprintf(blockSrvTag, process), v)
}

func getValidRetries(currentDate int64, oldRetries []int64, frequency string) []int64 {
	const oneDayMiliSeconds = 24 * 60 * 60 * 1000
	const oneMonthMiliSeconds = oneDayMiliSeconds * 30
	dateLimit := currentDate
	if strings.EqualFold(frequency, constants.DailyFrequency) {
		dateLimit = dateLimit - oneDayMiliSeconds
	} else {
		dateLimit = dateLimit - oneMonthMiliSeconds
	}
	retries := []int64{currentDate}
	for _, retry := range oldRetries {
		if retry > dateLimit {
			retries = append(retries, retry)
		}
	}

	return retries
}

func generateRetryKey(request types.BlockCardRequest, frequency string) string {
	customID := generateCustomID(request, frequency)

	return fmt.Sprintf("%s-%s", request.CardID, customID)
}

func generateCustomID(request types.BlockCardRequest, frequency string) string {
	if strings.EqualFold(request.Franchise, core.BrandMasterCard) {
		return fmt.Sprintf("%s-%s", request.MerchantIdentifier, frequency)
	} else {
		return fmt.Sprintf("%s-%s-%s", request.MerchantIdentifier, request.Conditional, frequency)
	}
}
