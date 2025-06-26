package adapters

import (
	"context"
	"fmt"

	"bitbucket.org/kushki/usrv-card-control/features/card-info/application/use_cases"
	"bitbucket.org/kushki/usrv-card-control/features/card-info/interfaces"
	"bitbucket.org/kushki/usrv-go-core/logger"
	"github.com/aws/aws-lambda-go/events"
)

// SQSAdapter implements MessageHandler for SQS events
type SQSAdapter struct {
	processCardInfoUseCase *use_cases.ProcessCardInfoMessageUseCase
	logger                 logger.KushkiLogger
}

// NewSQSAdapter creates a new SQS adapter that implements MessageHandler
func NewSQSAdapter(
	processCardInfoUseCase *use_cases.ProcessCardInfoMessageUseCase,
	logger logger.KushkiLogger,
) interfaces.MessageHandler {
	return &SQSAdapter{
		processCardInfoUseCase: processCardInfoUseCase,
		logger:                 logger,
	}
}

// ProcessCardInfoMessage implements the MessageHandler interface
func (a *SQSAdapter) ProcessCardInfoMessage(ctx context.Context, messageBody string) error {
	// Create use case request
	useCaseRequest := use_cases.ProcessCardInfoMessageRequest{
		SQSMessageBody: messageBody,
	}

	// Execute the use case
	response, err := a.processCardInfoUseCase.Execute(ctx, useCaseRequest)
	if err != nil {
		return fmt.Errorf("use case execution failed: %w", err)
	}

	a.logger.Info("SQSAdapter.ProcessCardInfoMessage | Success",
		fmt.Sprintf("ExternalReferenceID: %s, ProcessedAt: %d",
			response.ExternalReferenceID, response.ProcessedAt))

	return nil
}

// HandleSQSEvent processes SQS events (SQS-specific method)
func (a *SQSAdapter) HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	const adapter = "SQSAdapter.HandleSQSEvent"

	a.logger.Info(fmt.Sprintf("%s | Starting", adapter),
		fmt.Sprintf("Processing %d records", len(event.Records)))

	for i, record := range event.Records {
		a.logger.Info(fmt.Sprintf("%s | ProcessingRecord", adapter),
			fmt.Sprintf("Record %d/%d, MessageId: %s", i+1, len(event.Records), record.MessageId))

		// Use the interface method to process the message
		if err := a.ProcessCardInfoMessage(ctx, record.Body); err != nil {
			a.logger.Error(fmt.Sprintf("%s | RecordError", adapter),
				fmt.Sprintf("Failed to process record %d: %v", i+1, err))

			// In SQS, if any record fails, the entire batch should be retried
			return fmt.Errorf("failed to process SQS record %d (MessageId: %s): %w",
				i+1, record.MessageId, err)
		}

		a.logger.Info(fmt.Sprintf("%s | RecordSuccess", adapter),
			fmt.Sprintf("Successfully processed record %d/%d", i+1, len(event.Records)))
	}

	a.logger.Info(fmt.Sprintf("%s | Success", adapter),
		fmt.Sprintf("Successfully processed all %d records", len(event.Records)))

	return nil
}
