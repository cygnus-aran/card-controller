package card_info_processor_dlq_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	constants "bitbucket.org/kushki/usrv-card-control"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
	"github.com/rollbar/rollbar-go"
)

func RunCardInfoDLQ(_ context.Context, event events.SQSEvent) (bool, error) {
	const cardInfoDLQServiceTag = "cardInfoDLQ | %s"

	kskLogger, err := logger.NewKushkiLogger()
	if err != nil {
		return false, err
	}

	source := fmt.Sprintf(cardInfoDLQServiceTag, "NotifyRollbar")
	kskLogger.Info(source, event)

	rollbar.SetToken(os.Getenv(constants.EnvRollbarToken))
	rollbar.SetEnvironment(os.Getenv(constants.EnvUsrvStage))
	rollbar.SetCodeVersion(os.Getenv(constants.EnvUsrvCommit))
	rollbar.SetServerRoot(os.Getenv(constants.EnvUsrvStage))

	var request entities.CardInfoMessage
	if err = json.Unmarshal([]byte(event.Records[0].Body), &request); err != nil {
		return false, err
	}

	errorMessage := fmt.Errorf("error processing card info for externalReferenceId %s and merchant %s",
		request.ExternalReferenceID, request.MerchantID)
	kskLogger.Error(source, errorMessage.Error())
	rollbar.Error(errorMessage)
	rollbar.Wait()

	return true, nil
}

func main() {
	m := vesper.New(RunCardInfoDLQ).
		Use(middleware.InputOutputLogsMiddleware()).
		Use(middleware.DynamoParamsMiddleware(false))

	m.Start()
}
