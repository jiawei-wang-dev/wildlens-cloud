package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/repository"
)

var (
	ErrSpeciesRequired      = errors.New("species is required")
	ErrTagsRequired         = errors.New("at least one tag is required")
	ErrInvalidMinimumCount  = errors.New("minimum tag count must be greater than zero")
	ErrThumbnailURLRequired = errors.New("thumbnail_url is required")
	ErrURLsRequired         = errors.New("at least one URL is required")
	ErrTagOperationRequired = errors.New("operation is required")
)

// QueryService contains media query business logic.
type QueryService struct {
	repo repository.MediaRepository
}

func NewQueryService(repo repository.MediaRepository) *QueryService {
	return &QueryService{repo: repo}
}

func (s *QueryService) FindBySpecies(
	ctx context.Context,
	species string,
) ([]model.MediaFile, error) {
	species = strings.TrimSpace(species)

	if species == "" {
		return nil, ErrSpeciesRequired
	}

	return s.repo.FindBySpecies(ctx, species)
}

func (s *QueryService) FindByTagCounts(
	ctx context.Context,
	required map[string]int,
) ([]model.MediaFile, error) {
	if len(required) == 0 {
		return nil, ErrTagsRequired
	}

	for _, count := range required {
		if count <= 0 {
			return nil, ErrInvalidMinimumCount
		}
	}

	return s.repo.FindByTagCounts(ctx, required)
}

func (s *QueryService) FindOriginalByThumbnailURL(
	ctx context.Context,
	thumbnailURL string,
) (string, error) {
	thumbnailURL = strings.TrimSpace(thumbnailURL)

	if thumbnailURL == "" {
		return "", ErrThumbnailURLRequired
	}

	return s.repo.FindOriginalByThumbnailURL(ctx, thumbnailURL)
}

// UpdateTags validates and applies a bulk tag update.
func (s *QueryService) UpdateTags(
	ctx context.Context,
	urls []string,
	tags []string,
	operation *int,
) ([]model.MediaFile, error) {
	cleanURLs := make([]string, 0, len(urls))

	for _, url := range urls {
		url = strings.TrimSpace(url)

		if url != "" {
			cleanURLs = append(cleanURLs, url)
		}
	}

	if len(cleanURLs) == 0 {
		return nil, ErrURLsRequired
	}

	cleanTags := make([]string, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)

		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	if len(cleanTags) == 0 {
		return nil, ErrTagsRequired
	}

	if operation == nil {
		return nil, ErrTagOperationRequired
	}

	if *operation != repository.TagOperationRemove &&
		*operation != repository.TagOperationAdd {
		return nil, repository.ErrInvalidTagOperation
	}

	return s.repo.UpdateTags(
		ctx,
		cleanURLs,
		cleanTags,
		*operation,
	)
}
