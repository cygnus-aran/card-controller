package tools

import (
	"context"
	"errors"
	"testing"

	awsConf "bitbucket.org/kushki/usrv-card-control/config/aws"
	mocks "bitbucket.org/kushki/usrv-card-control/mocks/core"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

// TestInitializeDynamoClient tests cases for initialize Dynamo.
func TestInitializeDynamoClient(t *testing.T) {
	assertions := assert.New(t)
	lgg := &mocks.KushkiLogger{}
	t.Run("Initialize dynamo client successfully", func(t *testing.T) {
		ctx := context.Background()
		dynamoGtw, err := InitializeDynamoGtw(ctx, lgg)
		assertions.NotNil(dynamoGtw)
		assertions.Nil(err)
	})
	t.Run("Initialize dynamo client fails on awsConfig", func(t *testing.T) {
		ctx := context.Background()
		awsConfig = mockAwsProvideConfig(errors.New("error"))
		_, err := InitializeDynamoGtw(ctx, lgg)
		assertions.Error(err)
		t.Cleanup(resetMocks)
	})
}

func mockAwsProvideConfig(errorFake error) func(ctx context.Context, logger logger.KushkiLogger) (aws.Config, error) {
	return func(ctx context.Context, logger logger.KushkiLogger) (aws.Config, error) {
		return aws.Config{}, errorFake
	}
}

func resetMocks() {
	awsConfig = awsConf.ProvideAwsConfig
}
