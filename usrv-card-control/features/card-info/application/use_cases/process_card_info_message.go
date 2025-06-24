package use_cases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/repositories"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/services"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/value_objects"
	"bitbucket.org/kushki/usrv-go-core/logger"
)

// ProcessCardInfoMessageUseCase handles the complete business workflow for processing SQS card info messages
type ProcessCardInfoMessageUseCase struct {
	cardInfoRepo      repositories.CardInfoRepository
	encryptionService services.EncryptionService
	validationService services.ValidationService
	logger            logger.KushkiLogger
}

// NewProcessCardInfoMessageUseCase creates a new instance of the use case
func NewProcessCardInfoMessageUseCase(
	cardInfoRepo repositories.CardInfoRepository,
	encryptionService services.EncryptionService,
	validationService services.ValidationService,
	logger logger.KushkiLogger,
) *ProcessCardInfoMessageUseCase {
	return &ProcessCardInfoMessageUseCase{
		cardInfoRepo:      cardInfoRepo,
		encryptionService: encryptionService,
		validationService: validationService,
		logger:            logger,
	}
}

// ProcessCardInfoMessageRequest represents the input for the use case
type ProcessCardInfoMessageRequest struct {
	SQSMessageBody string
}

// ProcessCardInfoMessageResponse represents the output of the use case
type ProcessCardInfoMessageResponse struct {
	ExternalReferenceID string
	ProcessedAt         int64
	Success             bool
}

// Execute processes a card info message from SQS through the complete business workflow
func (uc *ProcessCardInfoMessageUseCase) Execute(
	ctx context.Context,
	request ProcessCardInfoMessageRequest,
) (*ProcessCardInfoMessageResponse, error) {
	const useCase = "ProcessCardInfoMessage"

	uc.logger.Info(fmt.Sprintf("%s | Starting", useCase), "Processing card info message")

	// Step 1: Parse the SQS message
	cardInfoMessage, err := uc.parseSQSMessage(request.SQSMessageBody)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("%s | ParseError", useCase), err)
		return nil, fmt.Errorf("failed to parse SQS message: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("%s | Parsed", useCase),
		fmt.Sprintf("ExternalReferenceID: %s, MerchantID: %s",
			cardInfoMessage.ExternalReferenceID, cardInfoMessage.MerchantID))

	// Step 2: Validate the message structure
	if err := uc.validateMessage(cardInfoMessage); err != nil {
		uc.logger.Error(fmt.Sprintf("%s | ValidationError", useCase), err)
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Step 3: Validate merchant access and credentials
	if err := uc.validateMerchantAccess(cardInfoMessage); err != nil {
		uc.logger.Error(fmt.Sprintf("%s | AccessError", useCase), err)
		return nil, fmt.Errorf("merchant access validation failed: %w", err)
	}

	// Step 4: Check if record already exists (idempotency)
	if exists, err := uc.checkIfAlreadyProcessed(ctx, cardInfoMessage.ExternalReferenceID); err != nil {
		uc.logger.Error(fmt.Sprintf("%s | ExistenceCheckError", useCase), err)
		return nil, fmt.Errorf("failed to check if already processed: %w", err)
	} else if exists {
		uc.logger.Info(fmt.Sprintf("%s | AlreadyProcessed", useCase),
			fmt.Sprintf("ExternalReferenceID: %s already exists", cardInfoMessage.ExternalReferenceID))
		return &ProcessCardInfoMessageResponse{
			ExternalReferenceID: cardInfoMessage.ExternalReferenceID,
			ProcessedAt:         time.Now().UnixMilli(),
			Success:             true,
		}, nil
	}

	// Step 5: Encrypt the card data
	encryptedCardData, err := uc.encryptCardData(cardInfoMessage)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("%s | EncryptionError", useCase), err)
		return nil, fmt.Errorf("failed to encrypt card data: %w", err)
	}

	// Step 6: Create the stored card info entity
	storedCardInfo := uc.createStoredCardInfo(cardInfoMessage, encryptedCardData)

	// Step 7: Save to DynamoDB
	if err := uc.saveCardInfo(ctx, storedCardInfo); err != nil {
		uc.logger.Error(fmt.Sprintf("%s | SaveError", useCase), err)
		return nil, fmt.Errorf("failed to save card info: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("%s | Success", useCase),
		fmt.Sprintf("Successfully processed ExternalReferenceID: %s", cardInfoMessage.ExternalReferenceID))

	return &ProcessCardInfoMessageResponse{
		ExternalReferenceID: cardInfoMessage.ExternalReferenceID,
		ProcessedAt:         time.Now().UnixMilli(),
		Success:             true,
	}, nil
}

// parseSQSMessage parses the SQS message body into a CardInfoMessage entity
func (uc *ProcessCardInfoMessageUseCase) parseSQSMessage(messageBody string) (*entities.CardInfoMessage, error) {
	var cardInfoMessage entities.CardInfoMessage

	if err := json.Unmarshal([]byte(messageBody), &cardInfoMessage); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	return &cardInfoMessage, nil
}

// validateMessage validates the basic message structure and required fields
func (uc *ProcessCardInfoMessageUseCase) validateMessage(message *entities.CardInfoMessage) error {
	if !message.IsValid() {
		return fmt.Errorf("message validation failed: missing required fields")
	}

	return uc.validationService.ValidateCardInfoMessage(message)
}

// validateMerchantAccess validates merchant access and private credentials
func (uc *ProcessCardInfoMessageUseCase) validateMerchantAccess(message *entities.CardInfoMessage) error {
	// Validate merchant has access to card info feature
	if err := uc.validationService.ValidateMerchantAccess(message.MerchantID); err != nil {
		return fmt.Errorf("merchant access denied: %w", err)
	}

	// Validate private credential
	if err := uc.validationService.ValidatePrivateCredential(
		message.PrivateCredentialID,
		message.MerchantID,
	); err != nil {
		return fmt.Errorf("invalid private credential: %w", err)
	}

	return nil
}

// checkIfAlreadyProcessed checks if the external reference ID has already been processed (idempotency)
func (uc *ProcessCardInfoMessageUseCase) checkIfAlreadyProcessed(ctx context.Context, externalReferenceID string) (bool, error) {
	_, err := uc.cardInfoRepo.FindByExternalReferenceID(ctx, externalReferenceID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// encryptCardData encrypts the card data using the merchant's public key
func (uc *ProcessCardInfoMessageUseCase) encryptCardData(message *entities.CardInfoMessage) (value_objects.EncryptedCardData, error) {
	encryptedData, err := uc.encryptionService.EncryptCardData(message.Card, message.MerchantID)
	if err != nil {
		return value_objects.EncryptedCardData{}, fmt.Errorf("encryption failed for merchant %s: %w", message.MerchantID, err)
	}

	return encryptedData, nil
}

// createStoredCardInfo creates a StoredCardInfo entity from the message and encrypted data
func (uc *ProcessCardInfoMessageUseCase) createStoredCardInfo(
	message *entities.CardInfoMessage,
	encryptedData value_objects.EncryptedCardData,
) *entities.StoredCardInfo {
	currentTime := time.Now().UnixMilli()
	expirationTime := time.Now().AddDate(0, 0, 180).UnixMilli() // 180 days from now

	return &entities.StoredCardInfo{
		ExternalReferenceID:  message.ExternalReferenceID,
		TransactionReference: message.TransactionReference,
		CardBrand:            message.CardBrand,
		TerminalID:           message.TerminalID,
		TransactionType:      message.TransactionType,
		TransactionStatus:    message.TransactionStatus,
		SubMerchantCode:      message.SubMerchantCode,
		IDAffiliation:        message.IDAffiliation,
		MerchantID:           message.MerchantID,
		PrivateCredentialID:  message.PrivateCredentialID,
		EncryptedCard:        encryptedData,
		TransactionDate:      currentTime,
		CreatedAt:            currentTime,
		ExpiresAt:            expirationTime,
	}
}

// saveCardInfo saves the card info to DynamoDB
func (uc *ProcessCardInfoMessageUseCase) saveCardInfo(ctx context.Context, cardInfo *entities.StoredCardInfo) error {
	return uc.cardInfoRepo.Save(ctx, cardInfo)
}
