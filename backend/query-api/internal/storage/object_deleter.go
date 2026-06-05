package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

const s3DeleteObjectsBatchSize = 1000

// MediaObjectDeleter removes original media and thumbnail objects.
type MediaObjectDeleter interface {
	DeleteMediaObjects(
		ctx context.Context,
		files []model.MediaFile,
	) error
}

// NoopObjectDeleter skips object deletion for local memory mode.
type NoopObjectDeleter struct{}

// NewNoopObjectDeleter creates a no-op media object deleter.
func NewNoopObjectDeleter() *NoopObjectDeleter {
	return &NoopObjectDeleter{}
}

// DeleteMediaObjects intentionally does nothing for local development.
func (d *NoopObjectDeleter) DeleteMediaObjects(
	_ context.Context,
	_ []model.MediaFile,
) error {
	return nil
}

// S3DeleteAPI contains the S3 operation required for deleting media objects.
type S3DeleteAPI interface {
	DeleteObjects(
		ctx context.Context,
		params *s3.DeleteObjectsInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectsOutput, error)
}

// S3ObjectDeleter removes media objects from S3.
type S3ObjectDeleter struct {
	client S3DeleteAPI
}

// NewS3ObjectDeleter creates an S3-backed media object deleter.
func NewS3ObjectDeleter(
	client S3DeleteAPI,
) *S3ObjectDeleter {
	return &S3ObjectDeleter{
		client: client,
	}
}

// DeleteMediaObjects removes originals and thumbnails grouped by bucket.
func (d *S3ObjectDeleter) DeleteMediaObjects(
	ctx context.Context,
	files []model.MediaFile,
) error {
	if len(files) == 0 {
		return nil
	}

	objectsByBucket, err := collectS3Objects(files)
	if err != nil {
		return err
	}

	for bucket, keys := range objectsByBucket {
		for start := 0; start < len(keys); start += s3DeleteObjectsBatchSize {
			end := start + s3DeleteObjectsBatchSize

			if end > len(keys) {
				end = len(keys)
			}

			if err := d.deleteObjectBatch(
				ctx,
				bucket,
				keys[start:end],
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *S3ObjectDeleter) deleteObjectBatch(
	ctx context.Context,
	bucket string,
	keys []string,
) error {
	objects := make([]types.ObjectIdentifier, 0, len(keys))

	for _, key := range keys {
		objects = append(
			objects,
			types.ObjectIdentifier{
				Key: aws.String(key),
			},
		)
	}

	output, err := d.client.DeleteObjects(
		ctx,
		&s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{
				Objects: objects,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("delete S3 objects from bucket %s: %w", bucket, err)
	}

	if output != nil && len(output.Errors) > 0 {
		firstError := output.Errors[0]

		return fmt.Errorf(
			"delete S3 object %s/%s: %s %s",
			bucket,
			aws.ToString(firstError.Key),
			aws.ToString(firstError.Code),
			aws.ToString(firstError.Message),
		)
	}

	return nil
}

func collectS3Objects(
	files []model.MediaFile,
) (map[string][]string, error) {
	results := make(map[string][]string)
	seenByBucket := make(map[string]map[string]struct{})

	for _, file := range files {
		bucket := strings.TrimSpace(file.Bucket)
		objectPath := file.ObjectPath
		thumbnailObjectPath := file.ThumbnailObjectPath

		if bucket == "" {
			return nil, ErrBucketRequired
		}

		if strings.TrimSpace(objectPath) == "" {
			return nil, ErrObjectPathRequired
		}

		addS3Object(results, seenByBucket, bucket, objectPath)

		if strings.TrimSpace(thumbnailObjectPath) != "" {
			addS3Object(
				results,
				seenByBucket,
				bucket,
				thumbnailObjectPath,
			)
		}
	}

	if len(results) == 0 {
		return nil, errors.New("no S3 objects to delete")
	}

	return results, nil
}

func addS3Object(
	objectsByBucket map[string][]string,
	seenByBucket map[string]map[string]struct{},
	bucket string,
	key string,
) {
	if _, exists := seenByBucket[bucket]; !exists {
		seenByBucket[bucket] = make(map[string]struct{})
	}

	if _, exists := seenByBucket[bucket][key]; exists {
		return
	}

	seenByBucket[bucket][key] = struct{}{}
	objectsByBucket[bucket] = append(objectsByBucket[bucket], key)
}
