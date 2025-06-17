// Package aws.
package aws

import (
	"context"
	"os"

	"bitbucket.org/kushki/usrv-go-core/logger"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

type (
	loaderFn = func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (cfg aws.Config, err error)
)

var (
	cfgLoader = awsConfig.LoadDefaultConfig
	options   = func(o *awsConfig.LoadOptions) error {
		o.Region = os.Getenv("AWS_REGION")
		return nil
	}
)

// ProvideAwsConfig provides an AWS Config and an error to FX.
func ProvideAwsConfig(ctx context.Context, logger logger.KushkiLogger) (aws.Config, error) {
	cfg, err := cfgLoader(ctx, options)
	if err != nil {
		return aws.Config{}, err
	}

	// Add LogReq and LogResp aws client middlewares to configuration.
	cfg = middleware.AddLogReqRespMiddlewares(logger, cfg)
	return cfg, nil
}
