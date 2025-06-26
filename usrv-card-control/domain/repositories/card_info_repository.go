package repositories

import (
	"bitbucket.org/kushki/usrv-card-control/features/card-info/domain/entities"
	"context"
)

// CardInfoRepository defines the contract for card information persistence
type CardInfoRepository interface {
	// Save stores the card information in the database
	Save(ctx context.Context, cardInfo *entities.StoredCardInfo) error

	// FindByExternalReferenceID retrieves card information by external reference ID
	FindByExternalReferenceID(ctx context.Context, externalReferenceID string) (*entities.StoredCardInfo, error)

	// Delete removes card information (for cleanup/expiration)
	Delete(ctx context.Context, externalReferenceID string) error

	// FindExpiredRecords finds records that have exceeded the 180-day limit
	FindExpiredRecords(ctx context.Context, currentTime int64) ([]*entities.StoredCardInfo, error)
}
