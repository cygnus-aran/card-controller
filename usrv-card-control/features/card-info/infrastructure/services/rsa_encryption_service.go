package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// RSAEncryptionService implements encryption using RSA with merchant public keys
type RSAEncryptionService struct {
	keyProvider MerchantKeyProvider
	logger      logger.KushkiLogger
}

// MerchantKeyProvider defines the interface for retrieving merchant public keys
type MerchantKeyProvider interface {
	GetMerchantPublicKey(merchantID string) (string, error)
	HasMerchantKey(merchantID string) bool
}

// NewRSAEncryptionService creates a new RSA encryption service
func NewRSAEncryptionService(
	keyProvider MerchantKeyProvider,
	logger logger.KushkiLogger,
) services.EncryptionService {
	return &RSAEncryptionService{
		keyProvider: keyProvider,
		logger:      logger,
	}
}

// EncryptCardData encrypts the card PAN and date using the merchant's public key
func (s *RSAEncryptionService) EncryptCardData(
	cardData value_objects.CardData,
	merchantID string,
) (value_objects.EncryptedCardData, error) {
	const operation = "RSAEncryptionService.EncryptCardData"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// Get merchant's public key
	publicKeyPEM, err := s.keyProvider.GetMerchantPublicKey(merchantID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s | KeyRetrievalError", operation), err)
		return value_objects.EncryptedCardData{}, fmt.Errorf("failed to get public key for merchant %s: %w", merchantID, err)
	}

	// Parse the public key
	publicKey, err := s.parsePublicKey(publicKeyPEM)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s | KeyParsingError", operation), err)
		return value_objects.EncryptedCardData{}, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Encrypt PAN
	encryptedPan, err := s.encryptData(cardData.Pan, publicKey)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s | PANEncryptionError", operation), err)
		return value_objects.EncryptedCardData{}, fmt.Errorf("failed to encrypt PAN: %w", err)
	}

	// Encrypt Date
	encryptedDate, err := s.encryptData(cardData.Date, publicKey)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s | DateEncryptionError", operation), err)
		return value_objects.EncryptedCardData{}, fmt.Errorf("failed to encrypt date: %w", err)
	}

	s.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	return value_objects.EncryptedCardData{
		EncryptedPan:  encryptedPan,
		EncryptedDate: encryptedDate,
	}, nil
}

// DecryptCardData decrypts the card data (if needed for internal operations)
func (s *RSAEncryptionService) DecryptCardData(
	encryptedData value_objects.EncryptedCardData,
	merchantID string,
) (value_objects.CardData, error) {
	// Note: This would require private key access, which typically
	// the service doesn't have. Implementation depends on requirements.
	return value_objects.CardData{}, fmt.Errorf("decryption not implemented - service only has public keys")
}

// ValidateMerchantKey validates if the merchant has a registered public key
func (s *RSAEncryptionService) ValidateMerchantKey(merchantID string) error {
	const operation = "RSAEncryptionService.ValidateMerchantKey"

	if !s.keyProvider.HasMerchantKey(merchantID) {
		s.logger.Error(fmt.Sprintf("%s | KeyNotFound", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return fmt.Errorf("no public key registered for merchant: %s", merchantID)
	}

	return nil
}

// parsePublicKey parses a PEM-encoded RSA public key
func (s *RSAEncryptionService) parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return rsaPub, nil
}

// encryptData encrypts data using RSA public key and returns base64 encoded result
func (s *RSAEncryptionService) encryptData(data string, publicKey *rsa.PublicKey) (string, error) {
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("RSA encryption failed: %w", err)
	}

	// Return base64 encoded encrypted data
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
