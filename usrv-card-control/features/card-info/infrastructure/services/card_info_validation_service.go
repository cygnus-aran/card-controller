package services

import (
	"fmt"
	"strings"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// CardInfoValidationService implements business validation rules
type CardInfoValidationService struct {
	merchantAccessProvider MerchantAccessProvider
	credentialProvider     CredentialProvider
	logger                 logger.KushkiLogger
}

// MerchantAccessProvider defines the interface for checking merchant access
type MerchantAccessProvider interface {
	HasCardInfoAccess(merchantID string) bool
	IsActiveMerchant(merchantID string) bool
}

// CredentialProvider defines the interface for validating private credentials
type CredentialProvider interface {
	ValidatePrivateCredential(privateCredentialID, merchantID string) bool
}

// NewCardInfoValidationService creates a new validation service
func NewCardInfoValidationService(
	merchantAccessProvider MerchantAccessProvider,
	credentialProvider CredentialProvider,
	logger logger.KushkiLogger,
) services.ValidationService {
	return &CardInfoValidationService{
		merchantAccessProvider: merchantAccessProvider,
		credentialProvider:     credentialProvider,
		logger:                 logger,
	}
}

// ValidateCardInfoMessage validates the incoming SQS message
func (s *CardInfoValidationService) ValidateCardInfoMessage(message *entities.PxpCardInfoMessage) error {
	const operation = "CardInfoValidationService.ValidateCardInfoMessage"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("ExternalReferenceID: %s", message.ExternalReferenceID))

	// Basic field validation
	if err := s.validateRequiredFields(message); err != nil {
		return fmt.Errorf("required field validation failed: %w", err)
	}

	// Business rule validation
	if err := s.validateBusinessRules(message); err != nil {
		return fmt.Errorf("business rule validation failed: %w", err)
	}

	s.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("ExternalReferenceID: %s", message.ExternalReferenceID))

	return nil
}

// ValidateMerchantAccess validates if merchant has access to card info feature
func (s *CardInfoValidationService) ValidateMerchantAccess(merchantID string) error {
	const operation = "CardInfoValidationService.ValidateMerchantAccess"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// Check if merchant is active
	if !s.merchantAccessProvider.IsActiveMerchant(merchantID) {
		s.logger.Error(fmt.Sprintf("%s | InactiveMerchant", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return fmt.Errorf("merchant is not active: %s", merchantID)
	}

	// Check if merchant has card info access
	if !s.merchantAccessProvider.HasCardInfoAccess(merchantID) {
		s.logger.Error(fmt.Sprintf("%s | AccessDenied", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return fmt.Errorf("merchant does not have card info access: %s", merchantID)
	}

	s.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	return nil
}

// ValidatePrivateCredential validates the private credential ID
func (s *CardInfoValidationService) ValidatePrivateCredential(privateCredentialID, merchantID string) error {
	const operation = "CardInfoValidationService.ValidatePrivateCredential"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	if !s.credentialProvider.ValidatePrivateCredential(privateCredentialID, merchantID) {
		s.logger.Error(fmt.Sprintf("%s | InvalidCredential", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return fmt.Errorf("invalid private credential for merchant: %s", merchantID)
	}

	s.logger.Info(fmt.Sprintf("%s | Success", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	return nil
}

// validateRequiredFields validates that all required fields are present and valid
func (s *CardInfoValidationService) validateRequiredFields(message *entities.PxpCardInfoMessage) error {
	if message.ExternalReferenceID == "" {
		return fmt.Errorf("externalReferenceId is required")
	}

	if message.TransactionReference == "" {
		return fmt.Errorf("transactionReference is required")
	}

	if message.MerchantID == "" {
		return fmt.Errorf("merchant_id is required")
	}

	if message.PrivateCredentialID == "" {
		return fmt.Errorf("privateCredentialId is required")
	}

	if message.Card.Pan == "" {
		return fmt.Errorf("card.pan is required")
	}

	if message.Card.Date == "" {
		return fmt.Errorf("card.date is required")
	}

	return nil
}

// validateBusinessRules validates business-specific rules
func (s *CardInfoValidationService) validateBusinessRules(message *entities.PxpCardInfoMessage) error {
	// Validate PAN format (basic check - should be numeric and reasonable length)
	if err := s.validatePAN(message.Card.Pan); err != nil {
		return err
	}

	// Validate expiration date format
	if err := s.validateExpirationDate(message.Card.Date); err != nil {
		return err
	}

	// Validate card brand is known
	if err := s.validateCardBrand(message.CardBrand); err != nil {
		return err
	}

	return nil
}

// validatePAN validates the PAN format
func (s *CardInfoValidationService) validatePAN(pan string) error {
	// Remove any spaces or dashes
	cleanPAN := strings.ReplaceAll(strings.ReplaceAll(pan, " ", ""), "-", "")

	// Check length (13-19 digits for valid cards)
	if len(cleanPAN) < 13 || len(cleanPAN) > 19 {
		return fmt.Errorf("invalid PAN length: must be 13-19 digits")
	}

	// Check if all characters are digits
	for _, char := range cleanPAN {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid PAN format: must contain only digits")
		}
	}

	return nil
}

// validateExpirationDate validates the expiration date format
func (s *CardInfoValidationService) validateExpirationDate(date string) error {
	// Basic format validation - expecting MMYY or MM/YY
	cleanDate := strings.ReplaceAll(date, "/", "")

	if len(cleanDate) != 4 {
		return fmt.Errorf("invalid expiration date format: expected MMYY or MM/YY")
	}

	// Check if all characters are digits
	for _, char := range cleanDate {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid expiration date format: must contain only digits")
		}
	}

	return nil
}

// validateCardBrand validates the card brand
func (s *CardInfoValidationService) validateCardBrand(brand string) error {
	validBrands := []string{"VISA", "MASTERCARD", "AMEX", "DISCOVER", "DINERS", "JCB"}

	upperBrand := strings.ToUpper(brand)
	for _, validBrand := range validBrands {
		if upperBrand == validBrand {
			return nil
		}
	}

	return fmt.Errorf("invalid card brand: %s", brand)
}
