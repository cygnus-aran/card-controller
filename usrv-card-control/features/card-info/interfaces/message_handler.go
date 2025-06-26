// features/card-info/application/interfaces/message_handler.go
package interfaces

import (
	"context"
)

// MessageHandler defines the contract for processing card info messages
// Each adapter (SQS, SNS, DynamoDB Streams, etc.) implements this interface
type MessageHandler interface {
	// ProcessCardInfoMessage handles the business logic for processing card info
	// The implementation is technology-agnostic - it only cares about the message content
	ProcessCardInfoMessage(ctx context.Context, messageBody string) error
}
