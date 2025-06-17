package types

// BlockCardRequest request for card blocking or increment retry.
type BlockCardRequest struct {
	MerchantIdentifier string `json:"merchantIdentifier,omitempty"`
	Franchise          string `json:"brand,omitempty"`
	Operation          string `json:"operation,omitempty"`
	CardID             string `json:"cardId,omitempty"`
	Processor          string `json:"processor,omitempty"`
	Conditional        string `json:"conditional,omitempty"`
}

// CardRetry card retry save information per merchant, code and card.
type CardRetry struct {
	CardID     string  `json:"cardID" dynamodbav:"cardID"`
	MerchantID string  `json:"merchantID" dynamodbav:"merchantID"`
	RetryKey   string  `json:"retryKey" dynamodbav:"retryKey"`
	Retries    []int64 `json:"retries" dynamodbav:"retries"`
	TimeStamp  int64   `json:"timeStamp" dynamodbav:"timeStamp"`
}

// RestoreDailyRequest info to clean daily retries and lastRetry timestamp.
type RestoreDailyRequest struct {
	CardID     string `json:"cardId"`
	MerchantID string `json:"merchantId"`
}
