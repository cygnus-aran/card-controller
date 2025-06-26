package value_objects

// EncryptedCardData represents encrypted card information for storage/retrieval
type EncryptedCardData struct {
	EncryptedPan  string `json:"encPan" dynamodbav:"encPan"`
	EncryptedDate string `json:"encDate" dynamodbav:"encDate"`
}

// IsValid validates the EncryptedCardData
func (e EncryptedCardData) IsValid() bool {
	return e.EncryptedPan != "" && e.EncryptedDate != ""
}
