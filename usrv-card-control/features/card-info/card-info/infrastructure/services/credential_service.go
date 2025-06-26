package services

import (
	"fmt"
	"os"
	"strings"

	domainServices "bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// CredentialService implements the CredentialService interface
type CredentialService struct {
	logger logger.KushkiLogger
}

// NewCredentialService creates a new credential service
func NewCredentialService(logger logger.KushkiLogger) domainServices.CredentialService {
	return &CredentialService{
		logger: logger,
	}
}

// ValidatePrivateCredential validates the private credential ID for a merchant
func (s *CredentialService) ValidatePrivateCredential(privateCredentialID, merchantID string) bool {
	const operation = "CredentialService.ValidatePrivateCredential"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// Basic validation - check if credential ID is not empty
	if privateCredentialID == "" {
		s.logger.Error(fmt.Sprintf("%s | EmptyCredential", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return false
	}

	// Check environment variable for valid credentials
	validCredentials := os.Getenv("CARD_INFO_VALID_CREDENTIALS")
	if validCredentials != "" {
		credentials := strings.Split(validCredentials, ",")
		for _, credential := range credentials {
			if strings.TrimSpace(credential) == privateCredentialID {
				s.logger.Info(fmt.Sprintf("%s | CredentialValid", operation),
					fmt.Sprintf("MerchantID: %s", merchantID))
				return true
			}
		}
	}

	// Check merchant-specific credential pattern
	expectedPattern := fmt.Sprintf("PRIV_%s_", merchantID)
	if strings.HasPrefix(privateCredentialID, expectedPattern) {
		s.logger.Info(fmt.Sprintf("%s | PatternMatched", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return true
	}

	// In test/dev environment, use lenient validation
	if isTestEnvironment() {
		// Accept any credential that starts with "TEST_" or is longer than 10 characters
		if strings.HasPrefix(privateCredentialID, "TEST_") || len(privateCredentialID) > 10 {
			s.logger.Info(fmt.Sprintf("%s | TestEnvironmentValid", operation),
				fmt.Sprintf("MerchantID: %s", merchantID))
			return true
		}
	}

	// TODO: Implement actual credential validation
	// This could involve:
	// 1. Querying a credential database
	// 2. Calling a credential validation service
	// 3. Checking against a merchant's registered credentials
	// 4. Validating credential format and expiration

	s.logger.Error(fmt.Sprintf("%s | CredentialInvalid", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))
	return false
}
