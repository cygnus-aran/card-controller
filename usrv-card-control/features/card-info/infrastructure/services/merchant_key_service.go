package services

import (
	"fmt"
	"os"

	domainServices "bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// MerchantKeyService implements the MerchantKeyProvider interface
type MerchantKeyService struct {
	logger logger.KushkiLogger
}

// NewMerchantKeyService creates a new merchant key service
func NewMerchantKeyService(logger logger.KushkiLogger) domainServices.MerchantKeyService {
	return &MerchantKeyService{
		logger: logger,
	}
}

// GetMerchantPublicKey retrieves the public key for a merchant
func (s *MerchantKeyService) GetMerchantPublicKey(merchantID string) (string, error) {
	const operation = "MerchantKeyService.GetMerchantPublicKey"

	s.logger.Info(fmt.Sprintf("%s | Starting", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))

	// TODO: Implement actual key retrieval logic
	// This could involve:
	// 1. Calling an external service
	// 2. Querying a database
	// 3. Reading from AWS Parameter Store/Secrets Manager
	// 4. Reading from environment variables for testing

	// For now, check if there's a test key in environment
	envKey := fmt.Sprintf("MERCHANT_%s_PUBLIC_KEY", merchantID)
	if key := os.Getenv(envKey); key != "" {
		s.logger.Info(fmt.Sprintf("%s | FoundInEnv", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return key, nil
	}

	// Mock implementation for development/testing
	if isTestEnvironment() {
		s.logger.Info(fmt.Sprintf("%s | UsingMockKey", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return getMockPublicKey(), nil
	}

	s.logger.Error(fmt.Sprintf("%s | KeyNotFound", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))
	return "", fmt.Errorf("public key not found for merchant: %s", merchantID)
}

// HasMerchantKey checks if a merchant has a registered public key
func (s *MerchantKeyService) HasMerchantKey(merchantID string) bool {
	const operation = "MerchantKeyService.HasMerchantKey"

	// Check environment variable
	envKey := fmt.Sprintf("MERCHANT_%s_PUBLIC_KEY", merchantID)
	if os.Getenv(envKey) != "" {
		return true
	}

	// In test environment, assume all merchants have keys
	if isTestEnvironment() {
		s.logger.Info(fmt.Sprintf("%s | MockEnvironment", operation),
			fmt.Sprintf("MerchantID: %s", merchantID))
		return true
	}

	// TODO: Implement actual key existence check
	s.logger.Info(fmt.Sprintf("%s | KeyNotFound", operation),
		fmt.Sprintf("MerchantID: %s", merchantID))
	return false
}

// Helper functions

func isTestEnvironment() bool {
	stage := os.Getenv("USRV_STAGE")
	return stage == "dev" || stage == "test" || stage == "local"
}

func getMockPublicKey() string {
	// This is a mock RSA public key for testing purposes
	return `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2rKz7QrJ3ztC9JPQXR1l
YNaJy8LU8Q5O1F2J6V3P8V9J1K2L3M4N5O6P7Q8R9S0T1U2V3W4X5Y6Z7A8B9C0D
1E2F3G4H5I6J7K8L9M0N1O2P3Q4R5S6T7U8V9W0X1Y2Z3A4B5C6D7E8F9G0H1I2J
3K4L5M6N7O8P9Q0R1S2T3U4V5W6X7Y8Z9A0B1C2D3E4F5G6H7I8J9K0L1M2N3O4P
5Q6R7S8T9U0V1W2X3Y4Z5A6B7C8D9E0F1G2H3I4J5K6L7M8N9O0P1Q2R3S4T5U6V
7W8X9Y0Z1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S0T1U2V3W4X5Y6Z7A8B
9QIDAQAB
-----END PUBLIC KEY-----`
}
