package main

import (
	"context"

	"bitbucket.org/kushki/usrv-card-control/service"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"bitbucket.org/kushki/usrv-go-core/rollbar"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
)

func blockCardHandler(ctx context.Context, event events.SQSEvent) (bool, error) {
	return true, service.InitBlockService(ctx, event)
}

func main() {
	m := vesper.New(blockCardHandler).
		Use(rollbar.WrapRollbar()).
		Use(middleware.InputOutputLogsMiddleware())

	m.Start()
}
