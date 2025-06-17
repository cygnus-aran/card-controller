package tools

import (
	"context"

	"bitbucket.org/kushki/usrv-card-control/config/aws"
	"bitbucket.org/kushki/usrv-go-core/gateway/dynamo"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// Definition of functions methods for testing purposes.
var (
	awsConfig = aws.ProvideAwsConfig
)

// InitializeDynamoGtw Initialize dynamo client.
func InitializeDynamoGtw(ctx context.Context, logger logger.KushkiLogger) (dynamo.IDynamoGateway, error) {
	cfg, err := awsConfig(ctx, logger)

	dynamoClient := dynamo.NewDynamoClient(cfg)
	dynamoGtw := dynamo.NewDynamoGateway(logger, dynamoClient)
	return dynamoGtw, err
}
