package entities

import "bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"

// PxpCardInfoMessage represents the incoming SQS message structure
type PxpCardInfoMessage struct {
	Card                 value_objects.CardData `json:"card"`
	ExternalReferenceID  string                 `json:"externalReferenceId"`
	TransactionReference string                 `json:"transactionReference"`
	CardBrand            string                 `json:"card_brand"`
	TerminalID           string                 `json:"terminalId"`
	TransactionType      string                 `json:"transactionType"`
	TransactionStatus    string                 `json:"transaction_status"`
	SubMerchantCode      string                 `json:"sub_merchant_code"`
	IDAffiliation        string                 `json:"id_affiliation"`
	MerchantID           string                 `json:"merchant_id"`
	PrivateCredentialID  string                 `json:"privateCredentialId"`
}

// IsValid validates the PxpCardInfoMessage
func (c *PxpCardInfoMessage) IsValid() bool {
	return c.ExternalReferenceID != "" &&
		c.TransactionReference != "" &&
		c.CardBrand != "" &&
		c.TerminalID != "" &&
		c.MerchantID != "" &&
		c.TransactionType != "" &&
		c.TransactionStatus != "" &&
		c.Card.IsValid()
}
