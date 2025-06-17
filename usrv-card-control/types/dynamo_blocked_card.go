package types

// DynamoBlockedCard blocked card information.
type DynamoBlockedCard struct {
	CardID           string                     `json:"cardID" dynamodbav:"cardID"`
	TimeStamp        int64                      `json:"timeStamp" dynamodbav:"timeStamp"`
	BlockedMerchants map[string]BlockedMerchant `json:"blockedMerchants" dynamodbav:"blockedMerchants"`
}

// BlockedMerchant saves timestamp per merchant blocking duration.
type BlockedMerchant struct {
	ExpirationDate int64  `json:"expirationDate" dynamodbav:"expirationDate,omitempty"`
	BlockType      string `json:"blockType" dynamodbav:"blockType,omitempty"`
	LastRetry      int64  `json:"lastRetry" dynamodbav:"lastRetry,omitempty"`
}
