package adapters

import (
	"fmt"
	"os"

	constants "bitbucket.org/kushki/usrv-card-control"
)

// DynamoTableAdapter provides table name resolution for card info
type DynamoTableAdapter struct{}

// GetCardInfoTableName returns the DynamoDB table name for card info
func (d *DynamoTableAdapter) GetCardInfoTableName() string {
	tableName := os.Getenv("DYNAMO_CARD_INFO_TABLE")
	if tableName == "" {
		// Fallback to a default pattern if not set
		stage := os.Getenv(constants.EnvUsrvStage)
		if stage == "" {
			stage = "dev"
		}
		tableName = fmt.Sprintf("card-info-%s", stage)
	}
	return tableName
}
