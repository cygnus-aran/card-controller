package services

import (
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
)

// EncryptionService defines the contract for card data encryption/decryption
type EncryptionService interface {
	// EncryptCardData encrypts the card PAN and date using the merchant's public key
	EncryptCardData(cardData value_objects.CardData, merchantID string) (value_objects.EncryptedCardData, error)
}
