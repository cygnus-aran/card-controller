package aws

import (
	"context"
	"errors"
	"testing"

	"bitbucket.org/kushki/usrv-card-control/mocks"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
)

func TestProvideAwsConfig(t *testing.T) {
	defer resetVars()
	_ = options(&awsConfig.LoadOptions{}) // does nothing. only for coverage

	t.Run("no errors getting default aws config", func(t *testing.T) {
		cfgLoader = generateLoaderFn(nil)
		lgg := mocks.GetMockLogger(t)
		got, err := ProvideAwsConfig(context.Background(), lgg)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "", got.Region)
	})

	t.Run("there's an error getting default aws config", func(t *testing.T) {
		cfgLoader = generateLoaderFn(errors.New("some error"))
		lgg := mocks.GetMockLogger(t)
		got, err := ProvideAwsConfig(context.Background(), lgg)

		assert.Error(t, err)
		assert.NotNil(t, got)
	})
}

func resetVars() {
	cfgLoader = awsConfig.LoadDefaultConfig
}

func generateLoaderFn(err error) loaderFn {
	return func(context.Context, ...func(*awsConfig.LoadOptions) error) (aws.Config, error) {
		if err != nil {
			return aws.Config{}, err
		}

		return aws.Config{Region: ""}, nil
	}
}
