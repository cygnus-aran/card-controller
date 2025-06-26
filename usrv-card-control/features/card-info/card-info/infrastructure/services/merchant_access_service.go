package services

import (
	"fmt"
	"os"
	"strings"

	domainServices "bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// MerchantAccessService implements the MerchantAccessService interface
type MerchantAccessService struct {
	logger logger.KushkiLogger
}

// NewMerchantAccessService creates a new merchant access service
func NewMerchantAccessService(logger logger.KushkiLogger) domainServices.MerchantAccessService {
	return &MerchantAccessService{
		logger: logger,
	}
}

// HasCardInfoAccess checks if a merchant has access to the card info feature
func (s *MerchantAccessService) HasCardInfoAccess(merchantID string) bool {
	const operation = "MerchantAccessService.HasCardInfoAccess"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// Check environment variable for allowed merchants
	allowedMerchants := os.Getenv("CARD_INFO_ALLOWED_MERCHANTS")
	if allowedMerchants != "" {
		merchants := strings.Split(allowedMerchants, ",")
		for _, merchant := range merchants {
			if strings.TrimSpace(merchant) == merchantID {
				s.logger.Info(fmt.Sprintf("%s | AccessGranted", operation),
					fmt.Sprintf("MerchantID: %s", merchantID))
				return true
			}
		}
	}

	// In test/dev environment, allow all merchants
	if isTestEnvironment() {
		s.logger.Info(fmt.Sprintf("%s | TestEnvironmentAccess", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return true
	}

	// TODO: Implement actual access check
	// This could involve:
	// 1. Querying a merchant configuration database
	// 2. Checking feature flags
	// 3. Calling a merchant service

	s.logger.Info(fmt.Sprintf("%s | AccessDenied", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))
	return false
}

// IsActiveMerchant checks if a merchant is active
func (s *MerchantAccessService) IsActiveMerchant(merchantID string) bool {
	const operation = "MerchantAccessService.IsActiveMerchant"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// Check environment variable for inactive merchants
	inactiveMerchants := os.Getenv("CARD_INFO_INACTIVE_MERCHANTS")
	if inactiveMerchants != "" {
		merchants := strings.Split(inactiveMerchants, ",")
		for _, merchant := range merchants {
			if strings.TrimSpace(merchant) == merchantID {
				s.logger.Info(fmt.Sprintf("%s | MerchantInactive", operation),
					fmt.Sprintf("MerchantID: %s", merchantID))
				return false
			}
		}
	}

	// In test/dev environment, assume all merchants are active
	if isTestEnvironment() {
		s.logger.Info(fmt.Sprintf("%s | TestEnvironmentActive", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return true
	}

	// TODO: Implement actual merchant status check
	// This could involve:
	// 1. Querying a merchant database
	// 2. Calling a merchant service
	// 3. Checking merchant configuration

	// For now, assume all merchants are active unless explicitly marked inactive
	s.logger.Info(fmt.Sprintf("%s | AssumedActive", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))
	return true
}
