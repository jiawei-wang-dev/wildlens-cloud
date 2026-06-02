package repository

import (
	"context"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

// MediaRepository defines the database operations used by the query service.
type MediaRepository interface {
	FindBySpecies(
		ctx context.Context,
		species string,
	) ([]model.MediaFile, error)

	FindByTagCounts(
		ctx context.Context,
		required map[string]int,
	) ([]model.MediaFile, error)

	FindOriginalByThumbnailURL(
		ctx context.Context,
		thumbnailURL string,
	) (string, error)
}
