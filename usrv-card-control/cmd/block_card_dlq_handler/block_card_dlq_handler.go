package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
	"github.com/rollbar/rollbar-go"
)

func RunLockCardDLQ(_ context.Context, event events.SQSEvent) (bool, error) {
	const blockedCardDLQServiceTag = "blockedCardDLQ | %s"
	kskLogger, err := logger.NewKushkiLogger()
	if err != nil {
		return false, err
	}
	source := fmt.Sprintf(blockedCardDLQServiceTag, "NotifyRollbar")
	kskLogger.Info(source, event)
	if err != nil {
		return false, err
	}

	rollbar.SetToken(os.Getenv(constants.EnvRollbarToken))
	rollbar.SetEnvironment(os.Getenv(constants.EnvUsrvStage))
	rollbar.SetCodeVersion(os.Getenv(constants.EnvUsrvCommit))
	rollbar.SetServerRoot(os.Getenv(constants.EnvUsrvStage))

	var request types.BlockCardRequest
	if err = json.Unmarshal([]byte(event.Records[0].Body), &request); err != nil {
		return false, err
	}

	errorMessage := fmt.Errorf("error de bloqueo de tarjeta %s para la transaccion %s", request.Franchise, request.CardID)
	kskLogger.Error(source, errorMessage.Error())
	rollbar.Error(errorMessage)
	rollbar.Wait()

	return true, nil
}

func main() {
	m := vesper.New(RunLockCardDLQ).
		Use(middleware.InputOutputLogsMiddleware()).
		Use(middleware.DynamoParamsMiddleware(false))

	m.Start()
}
