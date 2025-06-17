// Package constants.
package constants

import core "bitbucket.org/kushki/usrv-go-core"

// Operations and frequencies.
const (
	BlockCardOperation = "block"
	RetryCardOperation = "retry"

	DailyFrequency   = "daily"
	MonthlyFrequency = "monthly"
)

var (
	// Frequencies franchise frequencies to be applied.
	Frequencies = map[string][]string{
		core.BrandVisa: {
			MonthlyFrequency,
		},
		core.BrandMasterCard: {
			DailyFrequency, MonthlyFrequency,
		},
	}

	// Limits max retries per frequency.
	Limits = map[string]map[string]int{
		core.BrandVisa: {
			MonthlyFrequency: 15,
		},
		core.BrandMasterCard: {
			MonthlyFrequency: 35,
			DailyFrequency:   7,
		},
	}
)

// Block and Retry card fields.
const (
	CardIdField         = "cardID"
	LastRetryField      = "lastRetry"
	BlockedMerchants    = "blockedMerchants"
	MerchantIDField     = "merchantID"
	RetryKeyField       = "retryKey"
	RetriesField        = "retries"
	CardIdMerchantIndex = "cardIdMerchantIndex"
	TimeStamp           = "timeStamp"

	DynamoBlockedCard = "DYNAMO_BLOCKED_CARD"
	DynamoCardRetry   = "DYNAMO_CARD_RETRY"
	DayHours          = 24
	MonthDays         = 30
	Daily             = "daily"
)

// BLOCK TYPE
const (
	TEMPORARY = "TEMPORARY"
	PERMANENT = "PERMANENT"
)

// Rollbar config.
const (
	EnvRollbarToken = "ROLLBAR_TOKEN"
	EnvUsrvCommit   = "USRV_COMMIT"
	EnvUsrvName     = "USRV_NAME"
	EnvUsrvStage    = "USRV_STAGE"
)
