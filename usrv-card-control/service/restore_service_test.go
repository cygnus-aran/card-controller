package service

import (
	"context"
	"testing"

	"bitbucket.org/kushki/usrv-card-control/mocks"
	coreMock "bitbucket.org/kushki/usrv-card-control/mocks/core"
	mockService "bitbucket.org/kushki/usrv-card-control/mocks/service"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRestoreService(t *testing.T) {
	t.Run("should no be empty", func(t *testing.T) {
		assert.NotEmpty(t, NewRestoreService(mocks.GetMockLogger(t), &coreMock.IDynamoGateway{}))
	})
}

func TestInitRestoreService(t *testing.T) {
	t.Run("should return an error if error init dynamo", func(t *testing.T) {
		t.Cleanup(clean)
		newKushkiLogger = func(context.Context) logger.KushkiLogger {
			return mocks.GetMockLogger(t)
		}
		initializeDynamoGtw = func(context.Context, logger.KushkiLogger) (dynamo.IDynamoGateway, error) {
			return &coreMock.IDynamoGateway{}, commonError
		}
		err := InitRestoreService(context.TODO(), fakeEvent)
		assert.Error(t, err)
	})

	t.Run("should be successfully if not error on init dependencies", func(t *testing.T) {
		t.Cleanup(clean)
		newKushkiLogger = func(context.Context) logger.KushkiLogger {
			return mocks.GetMockLogger(t)
		}
		RefNewRestoreService = func(logger.KushkiLogger, dynamo.IDynamoGateway) IRestoreService {
			srv := &mockService.IRestoreService{}
			srv.On("RestoreDailyRetries", mock.Anything, mock.Anything).Return(nil)
			return srv
		}
		err := InitRestoreService(context.TODO(), fakeEvent)
		assert.NoError(t, err)
	})
}

type restoreDynamoErrors struct {
	Get    error
	Update error
	Query  error
	Delete error
}

type restoreScenario struct {
	Name           string
	DynamoErrors   restoreDynamoErrors
	HasError       bool
	UnmarshalError error
}

func TestRestoreService_RestoreDailyRetries(t *testing.T) {
	scenarios := []restoreScenario{
		{
			Name: "should be successfully",
		},
		{
			Name:         "should return an error if query fails",
			HasError:     true,
			DynamoErrors: restoreDynamoErrors{Query: commonError},
		},
		{
			Name:         "should return an error if get fails",
			HasError:     true,
			DynamoErrors: restoreDynamoErrors{Get: commonError},
		},
		{
			Name:         "should return an error if Delete fails",
			HasError:     true,
			DynamoErrors: restoreDynamoErrors{Delete: commonError},
		},
		{
			Name:           "should return an error if unmarshal fails",
			HasError:       true,
			UnmarshalError: commonError,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			dynamoGtw := coreMock.IDynamoGateway{}
			dynamoGtw.On("Query", mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					out := args[2].(*[]types.CardRetry)
					*out = make([]types.CardRetry, 1)
				}).
				Return(scenario.DynamoErrors.Query)
			dynamoGtw.On("DeleteItem", mock.Anything, mock.Anything).
				Return(scenario.DynamoErrors.Delete)
			dynamoGtw.On("GetItem", mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					out := args[2].(*types.DynamoBlockedCard)
					*out = types.DynamoBlockedCard{}
				}).
				Return(scenario.DynamoErrors.Get)

			dynamoGtw.On("UpdateItem", mock.Anything, mock.Anything).
				Return(scenario.DynamoErrors.Update)
			srv := RestoreService{
				Logger: mocks.GetMockLogger(t),
				Dynamo: &dynamoGtw,
			}
			jsonUnmarshalCaller = func(_ []byte, v any) error {
				out := v.(*types.RestoreDailyRequest)
				*out = types.RestoreDailyRequest{}

				return scenario.UnmarshalError
			}
			err := srv.RestoreDailyRetries(context.TODO(), fakeEvent)
			if scenario.HasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
