package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/mocks"
	coreMock "bitbucket.org/kushki/usrv-card-control/mocks/core"
	mockService "bitbucket.org/kushki/usrv-card-control/mocks/service"
	"bitbucket.org/kushki/usrv-card-control/tools"
	"bitbucket.org/kushki/usrv-card-control/types"
	core "bitbucket.org/kushki/usrv-go-core"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	dynamoerror "bitbucket.org/kushki/usrv-go-core/gateway/dynamo/errors"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type blockCardScenario struct {
	Name           string
	Request        types.BlockCardRequest
	DynamoErrors   dynamoErrors
	HasError       bool
	ExceededRetry  bool
	UnmarshalError error
	NewBlockedCard bool
	FranchiseMC    bool
	EmptyCardID    bool
}

type dynamoErrors struct {
	Update     error
	GetRetry   error
	GetBlocked error
	UpdateCard error
	Put        error
}

var (
	commonError = errors.New("some error")
	fakeEvent   = events.SQSEvent{
		Records: []events.SQSMessage{{Body: ""}},
	}
	notExpiredRetry   = time.Now().UTC().Add(time.Hour).UnixMilli() + 10000
	notExpiredRetries = func() []int64 {
		retries := make([]int64, 0)
		limit := constants.Limits[core.BrandVisa][constants.MonthlyFrequency]
		for i := 0; i <= limit; i++ {
			retries = append(retries, notExpiredRetry)
		}
		return retries
	}()
)

func TestBlockService_ProcessBlock_DirectBlock(t *testing.T) {
	t.Run("should block immediately when operation is block", func(t *testing.T) {
		dynamoMock := &coreMock.IDynamoGateway{}
		dynamoMock.On("GetItem", mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				out := args[2].(*types.DynamoBlockedCard)
				*out = types.DynamoBlockedCard{}
			}).
			Return(nil)
		dynamoMock.On("UpdateItem", mock.Anything, mock.Anything).
			Return(nil).
			Once()
		srv := BlockService{
			Logger: mocks.GetMockLogger(t),
			Dynamo: dynamoMock,
		}

		jsonUnmarshalCaller = func(_ []byte, v any) error {
			out := v.(*types.BlockCardRequest)
			*out = types.BlockCardRequest{Operation: constants.BlockCardOperation, CardID: "foo"}

			return nil
		}

		result := srv.ProcessBlock(context.TODO(), fakeEvent)
		assert.NoError(t, result)
		dynamoMock.AssertExpectations(t)
	})
}

func TestBlockService_ProcessBlock_Retry(t *testing.T) {
	scenarios := []blockCardScenario{
		{
			Name:           "should return an error if unmarshal error",
			HasError:       true,
			UnmarshalError: commonError,
		},
		{
			Name: "should be successfully when retries does not exceed limit",
		},
		{
			Name:        "should do nothing when cardID is empty",
			EmptyCardID: true,
		},
		{
			Name:         "should return an error when create new blockedCard return an error",
			HasError:     true,
			DynamoErrors: dynamoErrors{Put: commonError, GetBlocked: dynamoerror.ErrItemNotFound},
		},
		{
			Name:          "should return an error if update last retry fails",
			DynamoErrors:  dynamoErrors{Update: commonError},
			ExceededRetry: true,
			HasError:      true,
		},
		{
			Name:          "should block card if after adding new retry exceeds retry limit",
			ExceededRetry: true,
		},
		{
			Name:         "should return an error if getting retry return an error",
			DynamoErrors: dynamoErrors{GetRetry: commonError},
			HasError:     true,
		},
		{
			Name:         "should return an error if getting blocked card return an error",
			DynamoErrors: dynamoErrors{GetBlocked: commonError},
			HasError:     true,
		},
		{
			Name:        "should return an error if blocking by retry return an error",
			FranchiseMC: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			testProcessBlock(t, scenario)
		})
	}
}

func TestGenerateCustomID(t *testing.T) {
	t.Run("should work for MASTERCARD request", func(t *testing.T) {
		res := generateCustomID(types.BlockCardRequest{Franchise: core.BrandMasterCard}, "")
		assert.NotEmpty(t, res)
	})
}

func TestGetValidRetries(t *testing.T) {
	t.Run("should return daily time stamps", func(t *testing.T) {
		currentDate := time.Now().UTC().UnixMilli()
		const oneDayMiliSeconds = 24 * 60 * 60 * 1000
		oneDayValid := currentDate - 1000
		oneDayExpired := currentDate - oneDayMiliSeconds - 1000

		retries := []int64{oneDayValid, oneDayExpired}
		res := getValidRetries(currentDate, retries, constants.DailyFrequency)
		assert.Equal(t, 2, len(res))
	})
}

func testProcessBlock(t *testing.T, scenario blockCardScenario) {
	t.Helper()
	scenario.Request.CardID = "someCardId"
	scenario.Request.Operation = constants.RetryCardOperation
	scenario.Request.Franchise = core.BrandVisa
	updateRetries := 1
	if scenario.EmptyCardID {
		scenario.Request.CardID = ""
	}
	if scenario.FranchiseMC {
		scenario.Request.Franchise = core.BrandMasterCard
		updateRetries = 2
	}
	jsonUnmarshalCaller = func(_ []byte, v any) error {
		out := v.(*types.BlockCardRequest)
		*out = scenario.Request

		return scenario.UnmarshalError
	}
	dynamoMock := coreMock.IDynamoGateway{}
	dynamoMock.On("GetItem", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			out := args[2].(*types.DynamoBlockedCard)
			*out = types.DynamoBlockedCard{}
		}).
		Return(scenario.DynamoErrors.GetBlocked).
		Once()
	dynamoMock.On("PutItem", mock.Anything, mock.Anything).
		Return(scenario.DynamoErrors.Put)
	dynamoMock.On("GetItem", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			out := args[2].(*types.CardRetry)
			retry := types.CardRetry{}
			if scenario.ExceededRetry {
				retry.Retries = notExpiredRetries
				retry.TimeStamp = 3
			}
			*out = retry
		}).
		Return(scenario.DynamoErrors.GetRetry)
	dynamoMock.On("UpdateItem", mock.Anything, mock.Anything).
		Return(scenario.DynamoErrors.Update).Times(updateRetries)
	dynamoMock.On("UpdateItem", mock.Anything, mock.Anything).
		Return(scenario.DynamoErrors.UpdateCard)

	srv := BlockService{
		Logger: mocks.GetMockLogger(t),
		Dynamo: &dynamoMock,
	}
	err := srv.ProcessBlock(context.TODO(), fakeEvent)
	if scenario.HasError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewBlockService(t *testing.T) {
	t.Run("should no be empty", func(t *testing.T) {
		assert.NotEmpty(t, NewBlockService(mocks.GetMockLogger(t), &coreMock.IDynamoGateway{}))
	})
}

func TestInitBlockService(t *testing.T) {
	t.Run("should return an error if error init dynamo", func(t *testing.T) {
		t.Cleanup(clean)
		newKushkiLogger = func(context.Context) logger.KushkiLogger {
			return mocks.GetMockLogger(t)
		}
		initializeDynamoGtw = func(context.Context, logger.KushkiLogger) (dynamo.IDynamoGateway, error) {
			return &coreMock.IDynamoGateway{}, commonError
		}
		err := InitBlockService(context.TODO(), fakeEvent)
		assert.Error(t, err)
	})

	t.Run("should be successfully if not error on init dependencies", func(t *testing.T) {
		t.Cleanup(clean)
		newKushkiLogger = func(context.Context) logger.KushkiLogger {
			return mocks.GetMockLogger(t)
		}
		RefNewBlockService = func(logger.KushkiLogger, dynamo.IDynamoGateway) IBlockService {
			srv := &mockService.IBlockService{}
			srv.On("ProcessBlock", mock.Anything, mock.Anything).Return(nil)
			return srv
		}
		err := InitBlockService(context.TODO(), fakeEvent)
		assert.NoError(t, err)
	})
}

func clean() {
	RefNewRestoreService = NewRestoreService
	RefNewBlockService = NewBlockService
	refNewCheckCardStatusService = NewCheckCardStatusService
	newKushkiLogger = middleware.GetLoggerFromContext
	initializeDynamoGtw = tools.InitializeDynamoGtw
	jsonUnmarshalCaller = json.Unmarshal
}
