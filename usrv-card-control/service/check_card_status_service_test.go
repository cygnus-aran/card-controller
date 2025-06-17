package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/mocks"
	mocksCore "bitbucket.org/kushki/usrv-card-control/mocks/core"
	mockService "bitbucket.org/kushki/usrv-card-control/mocks/service"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	mockIDMerchantEnable      = "cardEnabled"
	mockIDMerchantPermanently = "lockPermanently"
	mockIDMerchantTemporarily = "lockTemporarily"
	mockIDMerchantRetries     = "lockRetries"
)

var (
	checkCardStatusRequest = types.CheckCardStatusRequest{}
)

type initializeCheckCardStatusTests struct {
	name                  string
	jsonUnmarshalError    error
	initializeDynamoError error
	expectedError         string
}

func TestInitializeCheckCardStatus(t *testing.T) {
	t.Helper()
	tests := []initializeCheckCardStatusTests{
		{
			name: "Should run successfully when initialize service",
		},
		{
			name:                  "Should error when initialize dynamo service",
			initializeDynamoError: errors.New("dynamo service error"),
			expectedError:         "Code: E002 | Status Code: 500 | Message: Ha ocurrido un error inesperado. | Metadata: {\"Origin\":\"CheckCardStatusServiceInitializeCheckCardStatus\",\"Message\":\"dynamo service error\"}",
		},
		{
			name:               "Should error when unmarshal json request",
			jsonUnmarshalError: errors.New("unmarshal Json Error"),
			expectedError:      "Code: E001 | Status Code: 400 | Message: Cuerpo de la petición inválido. | Metadata: {\"Origin\":\"CheckCardStatusServiceInitializeCheckCardStatus\",\"Message\":\"unmarshal Json Error\"}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertions := assert.New(t)
			t.Cleanup(clean)
			newKushkiLogger = func(context.Context) logger.KushkiLogger {
				return mocks.GetMockLogger(t)
			}
			mockCheckService := mockService.ICheckCardStatusService{}
			mockCheckService.On("CheckCardStatus", mock.Anything).Return(types.CheckCardStatusResponse{}, nil)

			refNewCheckCardStatusService = func(context.Context, dynamo.IDynamoGateway, logger.KushkiLogger) ICheckCardStatusService {
				return &mockCheckService
			}

			jsonUnmarshalCaller = func(data []byte, v any) error {
				return test.jsonUnmarshalError
			}
			initializeDynamoGtw = func(context.Context, logger.KushkiLogger) (dynamo.IDynamoGateway, error) {
				return &mocksCore.IDynamoGateway{}, test.initializeDynamoError
			}

			body := gabs.Wrap(checkCardStatusRequest).Bytes()
			eventRequest := events.APIGatewayProxyRequest{
				Body: string(body),
			}

			res, err := InitializeCheckCardStatus(context.Background(), eventRequest)
			if test.expectedError != "" {
				assertions.EqualError(err, test.expectedError)
			} else {
				assertions.NoError(err)
				assertions.IsType(types.CheckCardStatusResponse{}, res)
			}
		})
	}
}

type checkCardStatusServiceTest struct {
	name                       string
	merchantIDTest             string
	expectedError              string
	emptyCardId                bool
	getItemErr                 error
	expectedCardStatusResponse types.CheckCardStatusResponse
}

func TestCheckCardStatus(t *testing.T) {
	tests := []checkCardStatusServiceTest{
		{
			name:                       "Should run successfully when card is not blocked",
			merchantIDTest:             mockIDMerchantEnable,
			expectedCardStatusResponse: types.CheckCardStatusResponse{},
		},
		{
			name:                       "Should omit Logic and return empty response when cardId is empty",
			merchantIDTest:             mockIDMerchantEnable,
			emptyCardId:                true,
			expectedCardStatusResponse: types.CheckCardStatusResponse{},
		},
		{
			name:           "Should run successfully when card is permanent blocked",
			merchantIDTest: mockIDMerchantPermanently,
			expectedCardStatusResponse: types.CheckCardStatusResponse{
				BlockType: constants.PERMANENT,
				Blocked:   true,
			},
		},
		{
			name:           "Should run successfully when card is temporary blocked",
			merchantIDTest: mockIDMerchantTemporarily,
			expectedCardStatusResponse: types.CheckCardStatusResponse{
				BlockType: constants.TEMPORARY,
				Blocked:   true,
			},
		},
		{
			name:           "Should run successfully when card only has retries",
			merchantIDTest: mockIDMerchantRetries,
			expectedCardStatusResponse: types.CheckCardStatusResponse{
				HasRetries: true,
			},
		},
		{
			name:                       "Should return error when get card blocked info from dynamo ",
			getItemErr:                 errors.New("error getting blocked card info"),
			expectedCardStatusResponse: types.CheckCardStatusResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockCheckCardStatusService := makeMockCheckCardStatusService(t, test)
			mockRequest := types.CheckCardStatusRequest{
				CardID:             "CardTest123",
				MerchantIdentifier: test.merchantIDTest,
			}
			if test.emptyCardId {
				mockRequest.CardID = ""
			}

			cardStatusResponse := mockCheckCardStatusService.CheckCardStatus(mockRequest)
			assert.Equal(t, test.expectedCardStatusResponse, cardStatusResponse)

			t.Cleanup(clean)
		})
	}
}

func makeMockCheckCardStatusService(t *testing.T, test checkCardStatusServiceTest) ICheckCardStatusService {
	t.Helper()

	currentDate := time.Now()
	ctx := context.Background()
	mockLogger := mocks.GetMockLogger(t)
	mockDynamo := &mocksCore.IDynamoGateway{}
	mockDynamo.On("GetItem", mock.Anything, mock.Anything, mock.Anything).
		Return(test.getItemErr).
		Run(func(args mock.Arguments) {
			arg := args.Get(2)
			res, _ := json.Marshal(types.DynamoBlockedCard{
				CardID:    "CardTest123",
				TimeStamp: currentDate.Add(36 * time.Hour).UnixMilli(),
				BlockedMerchants: map[string]types.BlockedMerchant{
					mockIDMerchantPermanently: {
						BlockType: constants.PERMANENT,
					},
					mockIDMerchantTemporarily: {
						BlockType:      constants.TEMPORARY,
						ExpirationDate: currentDate.Add(2 * time.Hour).UnixMilli(),
					},
					mockIDMerchantRetries: {
						ExpirationDate: currentDate.Add(-2 * time.Hour).UnixMilli(),
						LastRetry:      currentDate.UnixMilli(),
					},
				},
			})

			err := jsonUnmarshalCaller(res, &arg)
			if err != nil {
				t.Fatal(err.Error())
			}
		})
	return refNewCheckCardStatusService(ctx, mockDynamo, mockLogger)
}
