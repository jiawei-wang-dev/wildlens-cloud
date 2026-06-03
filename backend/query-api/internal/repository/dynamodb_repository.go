package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

// DynamoDBScanAPI contains the DynamoDB operation used by this repository.
// A small interface makes it easier to test with a fake client.
type DynamoDBScanAPI interface {
	Scan(
		ctx context.Context,
		params *dynamodb.ScanInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ScanOutput, error)
}

// DynamoDBRepository reads media metadata from DynamoDB.
type DynamoDBRepository struct {
	client    DynamoDBScanAPI
	tableName string
}

// Compile-time check: DynamoDBRepository must implement MediaRepository.
var _ MediaRepository = (*DynamoDBRepository)(nil)

// NewDynamoDBRepository creates a DynamoDB-backed repository.
func NewDynamoDBRepository(
	client DynamoDBScanAPI,
	tableName string,
) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:    client,
		tableName: strings.TrimSpace(tableName),
	}
}

// FindBySpecies returns files containing at least one requested species.
func (r *DynamoDBRepository) FindBySpecies(
	ctx context.Context,
	species string,
) ([]model.MediaFile, error) {
	files, err := r.scanAll(ctx)
	if err != nil {
		return nil, err
	}

	species = strings.ToLower(strings.TrimSpace(species))
	results := make([]model.MediaFile, 0)

	for _, file := range files {
		if file.TagCounts[species] >= 1 {
			results = append(results, file)
		}
	}

	return results, nil
}

// FindByTagCounts applies logical AND between all minimum-count conditions.
func (r *DynamoDBRepository) FindByTagCounts(
	ctx context.Context,
	required map[string]int,
) ([]model.MediaFile, error) {
	files, err := r.scanAll(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]model.MediaFile, 0)

	for _, file := range files {
		matched := true

		for tag, minimumCount := range required {
			tag = strings.ToLower(strings.TrimSpace(tag))

			if file.TagCounts[tag] < minimumCount {
				matched = false
				break
			}
		}

		if matched {
			results = append(results, file)
		}
	}

	return results, nil
}

// FindOriginalByThumbnailURL maps a thumbnail URL to its original file URL.
func (r *DynamoDBRepository) FindOriginalByThumbnailURL(
	ctx context.Context,
	thumbnailURL string,
) (string, error) {
	files, err := r.scanAll(ctx)
	if err != nil {
		return "", err
	}

	thumbnailURL = strings.TrimSpace(thumbnailURL)

	for _, file := range files {
		if file.ThumbnailURL == thumbnailURL {
			return file.FileURL, nil
		}
	}

	return "", ErrMediaNotFound
}

// scanAll reads all DynamoDB Scan pages and converts items into MediaFile values.
func (r *DynamoDBRepository) scanAll(
	ctx context.Context,
) ([]model.MediaFile, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	paginator := dynamodb.NewScanPaginator(r.client, input)
	files := make([]model.MediaFile, 0)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan DynamoDB table: %w", err)
		}

		var pageFiles []model.MediaFile

		if err := attributevalue.UnmarshalListOfMaps(
			output.Items,
			&pageFiles,
		); err != nil {
			return nil, fmt.Errorf("decode DynamoDB items: %w", err)
		}

		files = append(files, pageFiles...)
	}

	return files, nil
}
