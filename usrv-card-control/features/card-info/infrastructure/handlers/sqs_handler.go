package handlers

import (
	"bitbucket.org/kushki/usrv-card-control/features/card-info/application/use_cases"
	"context"
	"fmt"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/infrastructure/config"
	"github.com/aws/aws-lambda-go/events"
)

// SQSCardInfoHandler handles SQS events for card info processing
type SQSCardInfoHandler struct {
	dependencies *config.DependencyContainer
}

// NewSQSCardInfoHandler creates a new SQS handler
func NewSQSCardInfoHandler(dependencies *config.DependencyContainer) *SQSCardInfoHandler {
	return &SQSCardInfoHandler{
		dependencies: dependencies,
	}
}

// HandleSQSEvent processes SQS events containing card information
func (h *SQSCardInfoHandler) HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	const handler = "SQSCardInfoHandler.HandleSQSEvent"

	h.dependencies.Logger.Info(fmt.Sprintf("%s | Starting", handler),
		fmt.Sprintf("Processing %d records", len(event.Records)))

	for i, record := range event.Records {
		h.dependencies.Logger.Info(fmt.Sprintf("%s | ProcessingRecord", handler),
			fmt.Sprintf("Record %d/%d, MessageId: %s", i+1, len(event.Records), record.MessageId))

		if err := h.processRecord(ctx, record); err != nil {
			h.dependencies.Logger.Error(fmt.Sprintf("%s | RecordError", handler),
				fmt.Sprintf("Failed to process record %d: %v", i+1, err))

			// In SQS, if any record fails, the entire batch should be retried
			// So we return the error to trigger retry
			return fmt.Errorf("failed to process SQS record %d (MessageId: %s): %w",
				i+1, record.MessageId, err)
		}

		h.dependencies.Logger.Info(fmt.Sprintf("%s | RecordSuccess", handler),
			fmt.Sprintf("Successfully processed record %d/%d", i+1, len(event.Records)))
	}

	h.dependencies.Logger.Info(fmt.Sprintf("%s | Success", handler),
		fmt.Sprintf("Successfully processed all %d records", len(event.Records)))

	return nil
}

// processRecord processes a single SQS record
func (h *SQSCardInfoHandler) processRecord(ctx context.Context, record events.SQSMessage) error {
	// Create the use case request
	request := use_cases.ProcessCardInfoMessageRequest{
		SQSMessageBody: record.Body,
	}

	// Execute the use case
	response, err := h.dependencies.ProcessCardInfoUseCase.Execute(ctx, request)
	if err != nil {
		return fmt.Errorf("use case execution failed: %w", err)
	}

	h.dependencies.Logger.Info("SQSCardInfoHandler.processRecord | UseCaseSuccess",
		fmt.Sprintf("ExternalReferenceID: %s, ProcessedAt: %d",
			response.ExternalReferenceID, response.ProcessedAt))

	return nil
}
