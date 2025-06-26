package services

// MerchantKeyService defines the interface for retrieving merchant public keys
type MerchantKeyService interface {
	GetMerchantPublicKey(merchantID string) (string, error)
	HasMerchantKey(merchantID string) bool
}
