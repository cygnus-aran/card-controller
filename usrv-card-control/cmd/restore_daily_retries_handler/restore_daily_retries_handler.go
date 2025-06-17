package main

import (
	"context"

	"bitbucket.org/kushki/usrv-card-control/service"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"bitbucket.org/kushki/usrv-go-core/rollbar"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
)

func restoreDailyRetriesHandler(ctx context.Context, event events.SQSEvent) (bool, error) {
	return true, service.InitRestoreService(ctx, event)
}

func main() {
	m := vesper.New(restoreDailyRetriesHandler).
		Use(rollbar.WrapRollbar()).
		Use(middleware.InputOutputLogsMiddleware())

	m.Start()
}
