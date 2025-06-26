package services

import (
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
)

// ValidationService defines the contract for business validation rules
type ValidationService interface {
	// ValidateCardInfoMessage validates the incoming SQS message
	ValidateCardInfoMessage(message *entities.PxpCardInfoMessage) error

	// ValidateMerchantAccess validates if merchant has access to card info feature
	ValidateMerchantAccess(merchantID string) error

	// ValidatePrivateCredential validates the private credential ID
	ValidatePrivateCredential(privateCredentialID, merchantID string) error
}
