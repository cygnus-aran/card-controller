// Package service functions.
package service

import (
	"encoding/json"

	"bitbucket.org/kushki/usrv-card-control/tools"
	"bitbucket.org/kushki/usrv-go-core/middleware"
)

// Definition of functions methods.
var (
	newKushkiLogger     = middleware.GetLoggerFromContext
	initializeDynamoGtw = tools.InitializeDynamoGtw
	jsonUnmarshalCaller = json.Unmarshal
)
