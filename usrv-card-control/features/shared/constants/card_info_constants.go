package constants

// Environment variable names for card info feature
const (
	// DynamoDB table
	EnvCardInfoTable = "DYNAMO_CARD_INFO_TABLE"

	// External service endpoints
	EnvMerchantKeyServiceURL    = "MERCHANT_KEY_SERVICE_URL"
	EnvMerchantAccessServiceURL = "MERCHANT_ACCESS_SERVICE_URL"
	EnvCredentialServiceURL     = "CREDENTIAL_SERVICE_URL"
)

// DynamoDB constants
const (
	// Table configuration
	CardInfoTableTTLDays = 180

	// Index names (if needed)
	MerchantIDIndex = "merchantId-index"
	ExpiresAtIndex  = "expiresAt-index"
)

// Business constants
const (
	// Card brands
	CardBrandVisa       = "VISA"
	CardBrandMasterCard = "MASTERCARD"
	CardBrandAmex       = "AMEX"
	CardBrandDiscover   = "DISCOVER"
	CardBrandDiners     = "DINERS"
	CardBrandJCB        = "JCB"

	// Transaction types
	TransactionTypeCapture         = "capture"
	TransactionTypeCharge          = "charge"
	TransactionTypePreAuth         = "preAuth"
	TransactionTypeReAuthorization = "reAuthorization"

	// Transaction statuses
	TransactionStatusApproval = "APPROVAL"
	TransactionStatusDeclined = "DECLINED"
)

// Validation constants
const (
	MinPANLength         = 13
	MaxPANLength         = 19
	ExpirationDateLength = 4
)
