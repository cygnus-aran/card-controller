package main

import (
	"bitbucket.org/kushki/usrv-card-control/features/card-info/infrastructure/handlers"
	"context"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/infrastructure/config"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"bitbucket.org/kushki/usrv-go-core/rollbar"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
)

func cardInfoProcessorHandler(ctx context.Context, event events.SQSEvent) (bool, error) {
	// Initialize dependencies
	dependencies, err := config.NewDependencyContainer(ctx)
	if err != nil {
		return false, err
	}

	// Create handler
	handler := handlers.NewSQSCardInfoHandler(dependencies)

	// Process the SQS event
	err = handler.HandleSQSEvent(ctx, event)
	if err != nil {
		return false, err
	}

	return true, nil
}

func main() {
	m := vesper.New(cardInfoProcessorHandler).
		Use(rollbar.WrapRollbar()).
		Use(middleware.InputOutputLogsMiddleware())

	m.Start()
}
