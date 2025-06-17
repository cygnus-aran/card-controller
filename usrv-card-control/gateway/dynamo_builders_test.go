package gateway

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	typesDynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

const (
	mockMerchantID = "merchantID1"
	mockTableName  = "table-name-mock"
	mockTimeStamp  = 1742508191000
)

func TestPutBlockedCardBuilder(t *testing.T) {
	t.Setenv(constants.DynamoBlockedCard, mockTableName)
	item := types.DynamoBlockedCard{}
	res := PutBlockedCardBuilder(item)
	input, err := res.BuildInput()
	if err != nil {
		t.Fatal("error marshal item", err.Error())
	}
	itemMarshal, err := attributevalue.MarshalMap(item)
	expected := &dynamodb.PutItemInput{
		Item:      itemMarshal,
		TableName: aws.String(mockTableName),
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, input)
}

func TestUpdateLastRetryBuilder(t *testing.T) {
	t.Setenv(constants.DynamoBlockedCard, mockTableName)

	item := types.DynamoBlockedCard{}
	item.CardID = "cardID123"
	res := UpdateLastRetryBuilder(mockTimeStamp, mockMerchantID, item)
	build, err := res.BuildInput()

	expected := &dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv(constants.DynamoBlockedCard)),
		ExpressionAttributeValues: map[string]typesDynamo.AttributeValue{
			":0": &typesDynamo.AttributeValueMemberS{
				Value: mockMerchantID,
			},
			":1": &typesDynamo.AttributeValueMemberS{
				Value: strconv.Itoa(mockTimeStamp),
			},
		},
		ExpressionAttributeNames: map[string]string{
			"#0": fmt.Sprintf("%s.%s", constants.BlockedMerchants, mockMerchantID),
			"#1": constants.TimeStamp,
		},
		Key: map[string]typesDynamo.AttributeValue{
			constants.CardIdField: &typesDynamo.AttributeValueMemberS{
				Value: item.CardID,
			},
		},
		UpdateExpression: aws.String("SET #0 = :0, #1 = :1"),
	}

	assert.ObjectsAreEqual(expected, build)
	assert.NoError(t, err)
}

func TestUpdateBlockCardBuilder(t *testing.T) {
	t.Setenv(constants.DynamoBlockedCard, mockTableName)
	itemRequest := types.BlockCardRequest{}

	t.Run("should build UpdateBlockCardBuilder with block operation is RetryCardOperation", func(t *testing.T) {
		itemRequest.CardID = "cardID123"
		itemRequest.Operation = constants.RetryCardOperation
		itemRequest.MerchantIdentifier = mockMerchantID
		itemCardBlocked := types.DynamoBlockedCard{}
		assertUpdateBlockCardBuilder(t, itemRequest, itemCardBlocked)
	})

	t.Run("should build UpdateBlockCardBuilder with block operation is BlockCardOperation", func(t *testing.T) {
		itemRequest.CardID = "cardID854"
		itemRequest.MerchantIdentifier = mockMerchantID
		itemRequest.Operation = constants.BlockCardOperation
		itemCardBlocked := types.DynamoBlockedCard{}
		assertUpdateBlockCardBuilder(t, itemRequest, itemCardBlocked)
	})
}

func assertUpdateBlockCardBuilder(t *testing.T, itemRequest types.BlockCardRequest, itemCardBlocked types.DynamoBlockedCard) {
	res := UpdateBlockCardBuilder(itemRequest, itemCardBlocked)
	build, err := res.BuildInput()

	expected := &dynamodb.UpdateItemInput{
		TableName: aws.String(mockTableName),
		Key: map[string]typesDynamo.AttributeValue{
			constants.CardIdField: &typesDynamo.AttributeValueMemberS{
				Value: itemRequest.CardID,
			},
		},
		ExpressionAttributeValues: map[string]typesDynamo.AttributeValue{
			":0": &typesDynamo.AttributeValueMemberS{
				Value: mockMerchantID,
			},
			":1": &typesDynamo.AttributeValueMemberS{
				Value: strconv.Itoa(mockTimeStamp),
			},
		},
		ExpressionAttributeNames: map[string]string{
			"#0": fmt.Sprintf("%s.%s", constants.BlockedMerchants, mockMerchantID),
			"#1": constants.TimeStamp,
		},
		UpdateExpression: aws.String("SET #0 = :0, #1 = :1"),
	}

	assert.ObjectsAreEqual(expected, build)
	assert.NoError(t, err)
}

func TestIncrementRetryBuilder(t *testing.T) {
	t.Setenv(constants.DynamoCardRetry, mockTableName)

	itemCardRetry := types.CardRetry{}
	itemRequest := types.BlockCardRequest{}
	itemRequest.CardID = "cardID123"
	itemRequest.MerchantIdentifier = mockMerchantID
	const mockKey = "key123"
	mockRetries := []int64{mockTimeStamp}
	res := IncrementRetryBuilder(mockRetries, itemRequest, mockKey, itemCardRetry)
	build, err := res.BuildInput()

	expected := &dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv(constants.DynamoCardRetry)),
		ExpressionAttributeValues: map[string]typesDynamo.AttributeValue{
			":0": &typesDynamo.AttributeValueMemberS{
				Value: strconv.Itoa(mockTimeStamp),
			},
			":1": &typesDynamo.AttributeValueMemberS{
				Value: fmt.Sprint(mockRetries),
			},
			":2": &typesDynamo.AttributeValueMemberS{
				Value: itemRequest.CardID,
			},
			":3": &typesDynamo.AttributeValueMemberS{
				Value: itemRequest.MerchantIdentifier,
			},
		},
		ExpressionAttributeNames: map[string]string{
			"#0": constants.TimeStamp,
			"#1": constants.RetriesField,
			"#2": constants.CardIdField,
			"#3": constants.MerchantIDField,
		},
		Key: map[string]typesDynamo.AttributeValue{
			constants.RetryKeyField: &typesDynamo.AttributeValueMemberS{
				Value: mockKey,
			},
		},
		UpdateExpression: aws.String("SET #0 = :0, #1 = :1, #2 = :2, #3 = :3"),
	}

	assert.ObjectsAreEqual(expected, build)
	assert.NoError(t, err)
}

func TestGetRetryBuilder(t *testing.T) {
	t.Setenv(constants.DynamoCardRetry, mockTableName)
	assertions := assert.New(t)
	key := "key_123"

	input, err := GetRetryBuilder(key).BuildInput()
	expected := &dynamodb.GetItemInput{
		TableName: aws.String(mockTableName),
		Key: map[string]typesDynamo.AttributeValue{
			constants.RetryKeyField: &typesDynamo.AttributeValueMemberS{
				Value: key,
			},
		},
		ConsistentRead: aws.Bool(true),
	}
	assertions.Equal(expected, input)
	assertions.NoError(err)
}

func TestGetBlockedCardBuilder(t *testing.T) {
	t.Setenv(constants.DynamoBlockedCard, mockTableName)
	assertions := assert.New(t)
	cardID := "card_test_123"

	input, err := GetBlockedCardBuilder(cardID).BuildInput()
	expected := &dynamodb.GetItemInput{
		TableName: aws.String(mockTableName),
		Key: map[string]typesDynamo.AttributeValue{
			constants.CardIdField: &typesDynamo.AttributeValueMemberS{
				Value: cardID,
			},
		},
		ConsistentRead: aws.Bool(true),
	}
	assertions.Equal(expected, input)
	assertions.NoError(err)
}

func TestDeleteCardRetryBuilder(t *testing.T) {
	t.Setenv(constants.DynamoCardRetry, mockTableName)
	assertions := assert.New(t)
	mockKey := "key_123"

	input, err := DeleteCardRetryBuilder(mockKey).BuildInput()
	expected := &dynamodb.DeleteItemInput{
		TableName: aws.String(mockTableName),
		Key: map[string]typesDynamo.AttributeValue{
			constants.RetryKeyField: &typesDynamo.AttributeValueMemberS{
				Value: mockKey,
			},
		},
	}
	assertions.Equal(expected, input)
	assertions.NoError(err)
}

func TestQueryRetriesBuilder(t *testing.T) {
	assertions := assert.New(t)

	t.Setenv(constants.DynamoCardRetry, mockTableName)
	resp := QueryRetriesBuilder("cardID123", mockMerchantID)
	build, err := resp.BuildInput()

	expected := &dynamodb.QueryInput{
		TableName: aws.String(mockTableName),
		IndexName: aws.String(constants.CardIdMerchantIndex),
		ExpressionAttributeValues: map[string]typesDynamo.AttributeValue{
			":0": &typesDynamo.AttributeValueMemberS{Value: constants.Daily},
			":1": &typesDynamo.AttributeValueMemberS{Value: "cardID123"},
			":2": &typesDynamo.AttributeValueMemberS{Value: mockMerchantID},
		},
		ExpressionAttributeNames: map[string]string{
			"#0": constants.RetryKeyField,
			"#1": constants.CardIdField,
			"#2": constants.MerchantIDField,
		},
		FilterExpression:       aws.String("contains (#0, :0)"),
		KeyConditionExpression: aws.String("(#1 = :1) AND (#2 = :2)"),
	}

	assertions.Equal(expected, build)
	assertions.NoError(err)
}
