package entities

import "bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"

// StoredCardInfo represents the complete card information stored in the database
type StoredCardInfo struct {
	ExternalReferenceID  string                          `json:"externalReferenceId" dynamodbav:"externalReferenceId"`
	TransactionReference string                          `json:"transactionReference" dynamodbav:"transactionReference"`
	CardBrand            string                          `json:"cardBrand" dynamodbav:"cardBrand"`
	TerminalID           string                          `json:"terminalId" dynamodbav:"terminalId"`
	TransactionType      string                          `json:"transactionType" dynamodbav:"transactionType"`
	TransactionStatus    string                          `json:"transactionStatus" dynamodbav:"transactionStatus"`
	SubMerchantCode      string                          `json:"subMerchantCode" dynamodbav:"subMerchantCode"`
	IDAffiliation        string                          `json:"idAffiliation" dynamodbav:"idAffiliation"`
	MerchantID           string                          `json:"merchantId" dynamodbav:"merchantId"`
	PrivateCredentialID  string                          `json:"privateCredentialId" dynamodbav:"privateCredentialId"`
	EncryptedCard        value_objects.EncryptedCardData `json:"card" dynamodbav:"card"`
	TransactionDate      int64                           `json:"transactionDate" dynamodbav:"transactionDate"`
	CreatedAt            int64                           `json:"createdAt" dynamodbav:"createdAt"`
	ExpiresAt            int64                           `json:"expiresAt" dynamodbav:"expiresAt"`
}

// IsExpired checks if the stored card info has expired (180 days)
func (s *StoredCardInfo) IsExpired(currentTime int64) bool {
	return currentTime > s.ExpiresAt
}
