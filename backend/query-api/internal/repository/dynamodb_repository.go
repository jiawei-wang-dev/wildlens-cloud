package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

// DynamoDBAPI contains the DynamoDB operations used by this repository.
// A small interface makes it easier to test with a fake client.
type DynamoDBAPI interface {
	Scan(
		ctx context.Context,
		params *dynamodb.ScanInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ScanOutput, error)

	UpdateItem(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)

	DeleteItem(
		ctx context.Context,
		params *dynamodb.DeleteItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DeleteItemOutput, error)
}

// DynamoDBRepository reads and updates media metadata in DynamoDB.
type DynamoDBRepository struct {
	client    DynamoDBAPI
	tableName string
}

// Compile-time check: DynamoDBRepository must implement MediaRepository.
var _ MediaRepository = (*DynamoDBRepository)(nil)

// NewDynamoDBRepository creates a DynamoDB-backed repository.
func NewDynamoDBRepository(
	client DynamoDBAPI,
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
		if matchesThumbnailLookup(file, thumbnailURL) {
			return file.FileURL, nil
		}
	}

	return "", ErrMediaNotFound
}

// FindByURLs returns media metadata matching original or thumbnail URLs.
func (r *DynamoDBRepository) FindByURLs(
	ctx context.Context,
	urls []string,
) ([]model.MediaFile, error) {
	files, err := r.scanAll(ctx)
	if err != nil {
		return nil, err
	}

	return findMediaFilesByURLs(files, urls), nil
}

// UpdateTags adds or removes tags for media files matching the supplied URLs.
func (r *DynamoDBRepository) UpdateTags(
	ctx context.Context,
	urls []string,
	tags []string,
	operation int,
) ([]model.MediaFile, error) {
	if err := validateTagOperation(operation); err != nil {
		return nil, err
	}

	files, err := r.scanAll(ctx)
	if err != nil {
		return nil, err
	}

	targetURLs := newURLSet(urls)
	updatedFiles := make([]model.MediaFile, 0)

	for index := range files {
		if !matchesMediaURL(files[index], targetURLs) {
			continue
		}

		if err := applyTagUpdate(
			&files[index],
			tags,
			operation,
		); err != nil {
			return nil, err
		}

		if err := r.persistTagUpdate(ctx, files[index]); err != nil {
			return nil, err
		}

		updatedFiles = append(updatedFiles, files[index])
	}

	return updatedFiles, nil
}

// DeleteFiles removes DynamoDB records matching the supplied file IDs.
func (r *DynamoDBRepository) DeleteFiles(
	ctx context.Context,
	fileIDs []string,
) ([]model.MediaFile, error) {
	targetIDs := newFileIDSet(fileIDs)

	files, err := r.scanAll(ctx)
	if err != nil {
		return nil, err
	}

	deletedFiles := make([]model.MediaFile, 0)

	for _, file := range files {
		if _, exists := targetIDs[file.FileID]; !exists {
			continue
		}

		if err := r.persistMediaDeletion(ctx, file.FileID); err != nil {
			return nil, err
		}

		deletedFiles = append(deletedFiles, file)
	}

	return deletedFiles, nil
}

// persistMediaDeletion removes one metadata record from DynamoDB.
func (r *DynamoDBRepository) persistMediaDeletion(
	ctx context.Context,
	fileID string,
) error {
	fileID = strings.TrimSpace(fileID)

	if fileID == "" {
		return fmt.Errorf("delete DynamoDB record: file_id is required")
	}

	_, err := r.client.DeleteItem(
		ctx,
		&dynamodb.DeleteItemInput{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"file_id": &types.AttributeValueMemberS{
					Value: fileID,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("delete DynamoDB record: %w", err)
	}

	return nil
}

// persistTagUpdate updates only tag-related DynamoDB fields.
// Other media metadata remains unchanged.
func (r *DynamoDBRepository) persistTagUpdate(
	ctx context.Context,
	file model.MediaFile,
) error {
	if strings.TrimSpace(file.FileID) == "" {
		return fmt.Errorf("persist tag update: file_id is required")
	}

	tagsValue, err := attributevalue.Marshal(file.Tags)
	if err != nil {
		return fmt.Errorf("encode tags: %w", err)
	}

	tagCountsValue, err := attributevalue.Marshal(file.TagCounts)
	if err != nil {
		return fmt.Errorf("encode tag counts: %w", err)
	}

	updatedAtValue, err := attributevalue.Marshal(file.UpdatedAt)
	if err != nil {
		return fmt.Errorf("encode updated_at: %w", err)
	}

	_, err = r.client.UpdateItem(
		ctx,
		&dynamodb.UpdateItemInput{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"file_id": &types.AttributeValueMemberS{
					Value: file.FileID,
				},
			},
			UpdateExpression: aws.String(
				"SET #tags = :tags, #tag_counts = :tag_counts, #updated_at = :updated_at",
			),
			ExpressionAttributeNames: map[string]string{
				"#tags":       "tags",
				"#tag_counts": "tag_counts",
				"#updated_at": "updated_at",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tags":       tagsValue,
				":tag_counts": tagCountsValue,
				":updated_at": updatedAtValue,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("update DynamoDB tags: %w", err)
	}

	return nil
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

// ListObservations returns filtered and paginated DynamoDB records.
func (r *DynamoDBRepository) ListObservations(
	ctx context.Context,
	options ObservationListOptions,
) (ObservationPage, error) {
	files, err := r.scanAll(ctx)
	if err != nil {
		return ObservationPage{}, err
	}

	return paginateObservations(files, options)
}
