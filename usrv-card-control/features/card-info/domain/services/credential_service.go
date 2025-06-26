package services

// CredentialService defines the interface for validating private credentials
type CredentialService interface {
	ValidatePrivateCredential(privateCredentialID, merchantID string) bool
}
