// Build transactions lambda.
package main

import (
	"context"
	"net/http"

	"bitbucket.org/kushki/usrv-card-control/service"
	"bitbucket.org/kushki/usrv-card-control/types"
	"bitbucket.org/kushki/usrv-go-core/middleware"
	"bitbucket.org/kushki/usrv-go-core/rollbar"
	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
)

const required = "required"

func authorizationTransactionHandler(ctx context.Context, event events.APIGatewayProxyRequest) (types.CheckCardStatusResponse, error) {
	return service.InitializeCheckCardStatus(ctx, event)
}

func main() {

	baseRules := map[string]interface{}{
		"merchantIdentifier": required,
	}

	m := vesper.New(authorizationTransactionHandler).
		Use(rollbar.WrapRollbar()).
		Use(middleware.ErrorAPIMiddleware(false)).
		Use(middleware.InputOutputLogsMiddleware()).
		Use(middleware.SchemaValidationMiddleware(types.CheckCardStatusRequest{}, baseRules)).
		Use(middleware.APIGatewayMiddleware(middleware.ContentTypeJSON, middleware.ContentTypeJSON, true, http.StatusOK))
	m.Start()
}
