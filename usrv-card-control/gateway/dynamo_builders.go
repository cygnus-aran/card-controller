package gateway

import (
	"fmt"
	"os"
	"strings"
	"time"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo/builder"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

func PutBlockedCardBuilder(blockedCard types.DynamoBlockedCard) *builder.PutItemBuilder {
	return builder.NewPutItemBuilder().
		WithItem(blockedCard).
		WithTable(os.Getenv(constants.DynamoBlockedCard))
}

func UpdateLastRetryBuilder(currentDate int64, merchantID string, blockedCard types.DynamoBlockedCard) *builder.UpdateItemBuilder {
	newVersion := time.Now().UTC().UnixMilli()
	blockedMerchant := types.BlockedMerchant{
		LastRetry: currentDate,
	}
	update := expression.Set(
		expression.Name(fmt.Sprintf("%s.%s", constants.BlockedMerchants, merchantID)),
		expression.Value(blockedMerchant)).
		Set(expression.Name(constants.TimeStamp), expression.Value(newVersion))
	condition := expression.Name(constants.TimeStamp).Equal(expression.Value(blockedCard.TimeStamp)) // optimistic concurrency.
	expr := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition)

	return builder.NewUpdateItemBuilder().
		WithTable(os.Getenv(constants.DynamoBlockedCard)).
		WithPartitionKey(constants.CardIdField, blockedCard.CardID).
		WithExpression(&expr)
}

func IncrementRetryBuilder(retries []int64, request types.BlockCardRequest, key string, cardRetry types.CardRetry) *builder.UpdateItemBuilder {
	newVersion := time.Now().UTC().UnixMilli()

	update := expression.Set(expression.Name(constants.TimeStamp), expression.Value(newVersion)).
		Set(expression.Name(constants.RetriesField), expression.Value(retries)).
		Set(expression.Name(constants.CardIdField), expression.IfNotExists(expression.Name(constants.CardIdField), expression.Value(request.CardID))).
		Set(expression.Name(constants.MerchantIDField), expression.IfNotExists(expression.Name(constants.MerchantIDField), expression.Value(request.MerchantIdentifier)))
	condition := expression.Name(constants.TimeStamp).Equal(expression.Value(cardRetry.TimeStamp)).
		Or(expression.Name(constants.TimeStamp).AttributeNotExists()) // optimistic concurrency.
	expr := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition)

	return builder.NewUpdateItemBuilder().
		WithTable(os.Getenv(constants.DynamoCardRetry)).
		WithPartitionKey(constants.RetryKeyField, key).
		WithExpression(&expr)
}

func GetRetryBuilder(key string) *builder.GetItemBuilder {
	return builder.NewGetItemBuilder().
		WithTable(os.Getenv(constants.DynamoCardRetry)).
		WithPartitionKey(constants.RetryKeyField, key).
		WithConsistentRead(true)
}

func GetBlockedCardBuilder(cardID string) *builder.GetItemBuilder {
	return builder.NewGetItemBuilder().
		WithTable(os.Getenv(constants.DynamoBlockedCard)).
		WithPartitionKey(constants.CardIdField, cardID).
		WithConsistentRead(true)
}

func UpdateBlockCardBuilder(request types.BlockCardRequest, blockedCard types.DynamoBlockedCard) *builder.UpdateItemBuilder {
	update := generateBlockUpdate(request.MerchantIdentifier, request.Operation)
	condition := expression.Name(constants.TimeStamp).Equal(expression.Value(blockedCard.TimeStamp))

	exprBuilder := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition)

	return builder.NewUpdateItemBuilder().
		WithTable(os.Getenv(constants.DynamoBlockedCard)).
		WithPartitionKey(constants.CardIdField, request.CardID).
		WithExpression(&exprBuilder)
}

func generateBlockUpdate(merchantID string, operation string) expression.UpdateBuilder {
	var blockType string
	if strings.EqualFold(operation, constants.BlockCardOperation) {
		blockType = constants.PERMANENT
	} else {
		blockType = constants.TEMPORARY
	}
	newBlockedMerchant := types.BlockedMerchant{
		ExpirationDate: generateExpirationDate(),
		BlockType:      blockType,
	}
	newVersion := time.Now().UTC().UnixMilli()
	return expression.Set(
		expression.Name(fmt.Sprintf("%s.%s", constants.BlockedMerchants, merchantID)),
		expression.Value(newBlockedMerchant)).
		Set(expression.Name(constants.TimeStamp), expression.Value(newVersion))
}

func generateExpirationDate() int64 {
	timeToAdd := time.Hour * time.Duration(constants.DayHours)

	return time.Now().UTC().Add(timeToAdd).UnixMilli()
}

func DeleteCardRetryBuilder(key string) *builder.DeleteItemBuilder {
	return builder.NewDeleteItemBuilder().
		WithTable(os.Getenv(constants.DynamoCardRetry)).
		WithPartitionKey(constants.RetryKeyField, key)
}

func QueryRetriesBuilder(cardID string, merchantID string) *builder.QueryBuilder {
	keyCondition := expression.Key(constants.CardIdField).
		Equal(expression.Value(cardID)).
		And(expression.Key(constants.MerchantIDField).
			Equal(expression.Value(merchantID)))

	filter := expression.Name(constants.RetryKeyField).Contains(constants.Daily)

	expr := expression.NewBuilder().
		WithKeyCondition(keyCondition).
		WithFilter(filter)

	return builder.NewQueryBuilder().
		WithTable(os.Getenv(constants.DynamoCardRetry)).
		WithIndexName(constants.CardIdMerchantIndex).
		WithExpression(&expr)
}
