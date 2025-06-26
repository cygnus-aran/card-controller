package services

// MerchantAccessService defines the interface for checking merchant access
type MerchantAccessService interface {
	HasCardInfoAccess(merchantID string) bool
	IsActiveMerchant(merchantID string) bool
}
