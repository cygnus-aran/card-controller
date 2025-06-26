package config

import (
	"context"
	"fmt"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/application/use_cases"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/infrastructure/repositories"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/infrastructure/services"
	"bitbucket.org/kushki/usrv-card-control/tools"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// DependencyContainer holds all the dependencies for the card-info feature
type DependencyContainer struct {
	// Use Cases
	ProcessCardInfoUseCase *use_cases.ProcessCardInfoMessageUseCase

	// Infrastructure
	Logger logger.KushkiLogger
}

// NewDependencyContainer creates and wires up all dependencies
func NewDependencyContainer(ctx context.Context) (*DependencyContainer, error) {
	// Initialize logger
	kskLogger, err := logger.NewKushkiLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize DynamoDB gateway (reusing existing utility)
	dynamoGtw, err := tools.InitializeDynamoGtw(ctx, kskLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DynamoDB gateway: %w", err)
	}

	// Create repository
	cardInfoRepo := repositories.NewDynamoCardInfoRepository(dynamoGtw, kskLogger)

	// Create concrete service implementations
	keyProvider := services.NewMerchantKeyService(kskLogger)
	encryptionService := services.NewRSAEncryptionService(keyProvider, kskLogger)

	merchantAccessProvider := services.NewMerchantAccessService(kskLogger)
	credentialProvider := services.NewCredentialService(kskLogger)
	validationService := services.NewCardInfoValidationService(
		merchantAccessProvider,
		credentialProvider,
		kskLogger,
	)

	// Create use case
	processCardInfoUseCase := use_cases.NewProcessCardInfoMessageUseCase(
		cardInfoRepo,
		encryptionService,
		validationService,
		kskLogger,
	)

	return &DependencyContainer{
		ProcessCardInfoUseCase: processCardInfoUseCase,
		Logger:                 kskLogger,
	}, nil
}
