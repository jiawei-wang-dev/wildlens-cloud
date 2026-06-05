package storage

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

type fakeS3DeleteClient struct {
	inputs []*s3.DeleteObjectsInput
	err    error
	output *s3.DeleteObjectsOutput
}

func (f *fakeS3DeleteClient) DeleteObjects(
	_ context.Context,
	params *s3.DeleteObjectsInput,
	_ ...func(*s3.Options),
) (*s3.DeleteObjectsOutput, error) {
	if f.err != nil {
		return nil, f.err
	}

	f.inputs = append(f.inputs, params)

	if f.output != nil {
		return f.output, nil
	}

	return &s3.DeleteObjectsOutput{}, nil
}

func TestS3ObjectDeleterDeletesImageOriginalAndThumbnail(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:              "wildlens-media",
				ObjectPath:          "media/originals/koala.jpg",
				ThumbnailObjectPath: "media/thumbnails/koala.jpg",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(client.inputs) != 1 {
		t.Fatalf("expected 1 DeleteObjects call, got %d", len(client.inputs))
	}

	assertDeleteInputKeys(
		t,
		client.inputs[0],
		"wildlens-media",
		[]string{
			"media/originals/koala.jpg",
			"media/thumbnails/koala.jpg",
		},
	)
}

func TestS3ObjectDeleterDeletesVideoOriginalOnly(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:     "wildlens-media",
				ObjectPath: "media/originals/wombat.mp4",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(client.inputs) != 1 {
		t.Fatalf("expected 1 DeleteObjects call, got %d", len(client.inputs))
	}

	assertDeleteInputKeys(
		t,
		client.inputs[0],
		"wildlens-media",
		[]string{"media/originals/wombat.mp4"},
	)
}

func TestS3ObjectDeleterDeduplicatesKeysWithinBucket(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:              "wildlens-media",
				ObjectPath:          "media/originals/koala.jpg",
				ThumbnailObjectPath: "media/thumbnails/koala.jpg",
			},
			{
				Bucket:              "wildlens-media",
				ObjectPath:          "media/originals/koala.jpg",
				ThumbnailObjectPath: "media/thumbnails/koala.jpg",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertDeleteInputKeys(
		t,
		client.inputs[0],
		"wildlens-media",
		[]string{
			"media/originals/koala.jpg",
			"media/thumbnails/koala.jpg",
		},
	)
}

func TestS3ObjectDeleterGroupsDeletesByBucket(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:     "wildlens-images",
				ObjectPath: "media/originals/koala.jpg",
			},
			{
				Bucket:     "wildlens-videos",
				ObjectPath: "media/originals/wombat.mp4",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(client.inputs) != 2 {
		t.Fatalf("expected 2 DeleteObjects calls, got %d", len(client.inputs))
	}

	keysByBucket := keysByBucket(client.inputs)

	assertKeysEqual(
		t,
		keysByBucket["wildlens-images"],
		[]string{"media/originals/koala.jpg"},
	)
	assertKeysEqual(
		t,
		keysByBucket["wildlens-videos"],
		[]string{"media/originals/wombat.mp4"},
	)
}

func TestS3ObjectDeleterSplitsBatchesAtOneThousandKeys(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)
	files := make([]model.MediaFile, 0, 1001)

	for index := 0; index < 1001; index++ {
		files = append(
			files,
			model.MediaFile{
				Bucket:     "wildlens-media",
				ObjectPath: "media/originals/file-" + stringID(index),
			},
		)
	}

	err := deleter.DeleteMediaObjects(context.Background(), files)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(client.inputs) != 2 {
		t.Fatalf("expected 2 DeleteObjects calls, got %d", len(client.inputs))
	}

	if len(client.inputs[0].Delete.Objects) != 1000 {
		t.Fatalf(
			"expected first batch to contain 1000 keys, got %d",
			len(client.inputs[0].Delete.Objects),
		)
	}

	if len(client.inputs[1].Delete.Objects) != 1 {
		t.Fatalf(
			"expected second batch to contain 1 key, got %d",
			len(client.inputs[1].Delete.Objects),
		)
	}
}

func TestS3ObjectDeleterReturnsSDKError(t *testing.T) {
	client := &fakeS3DeleteClient{
		err: errors.New("delete failed"),
	}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:     "wildlens-media",
				ObjectPath: "media/originals/koala.jpg",
			},
		},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestS3ObjectDeleterReturnsDeleteObjectsOutputError(t *testing.T) {
	client := &fakeS3DeleteClient{
		output: &s3.DeleteObjectsOutput{
			Errors: []types.Error{
				{
					Key:     aws.String("media/originals/koala.jpg"),
					Code:    aws.String("AccessDenied"),
					Message: aws.String("access denied"),
				},
			},
		},
	}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:     "wildlens-media",
				ObjectPath: "media/originals/koala.jpg",
			},
		},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestS3ObjectDeleterRejectsMissingBucket(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				ObjectPath: "media/originals/koala.jpg",
			},
		},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if len(client.inputs) != 0 {
		t.Fatalf("expected no S3 calls, got %d", len(client.inputs))
	}
}

func TestS3ObjectDeleterRejectsMissingObjectPath(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket: "wildlens-media",
			},
		},
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if len(client.inputs) != 0 {
		t.Fatalf("expected no S3 calls, got %d", len(client.inputs))
	}
}

func TestS3ObjectDeleterSkipsEmptyFileList(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(context.Background(), []model.MediaFile{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(client.inputs) != 0 {
		t.Fatalf("expected no S3 calls, got %d", len(client.inputs))
	}
}

func TestS3ObjectDeleterPreservesObjectPathSpaces(t *testing.T) {
	client := &fakeS3DeleteClient{}
	deleter := NewS3ObjectDeleter(client)

	err := deleter.DeleteMediaObjects(
		context.Background(),
		[]model.MediaFile{
			{
				Bucket:     "fit5225-wildlife-media",
				ObjectPath: "true (1).jpg",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertDeleteInputKeys(
		t,
		client.inputs[0],
		"fit5225-wildlife-media",
		[]string{"true (1).jpg"},
	)
}

func assertDeleteInputKeys(
	t *testing.T,
	input *s3.DeleteObjectsInput,
	expectedBucket string,
	expectedKeys []string,
) {
	t.Helper()

	if aws.ToString(input.Bucket) != expectedBucket {
		t.Fatalf(
			"expected bucket %q, got %q",
			expectedBucket,
			aws.ToString(input.Bucket),
		)
	}

	keys := make([]string, 0, len(input.Delete.Objects))

	for _, object := range input.Delete.Objects {
		keys = append(keys, aws.ToString(object.Key))
	}

	assertKeysEqual(t, keys, expectedKeys)
}

func keysByBucket(
	inputs []*s3.DeleteObjectsInput,
) map[string][]string {
	results := make(map[string][]string)

	for _, input := range inputs {
		bucket := aws.ToString(input.Bucket)

		for _, object := range input.Delete.Objects {
			results[bucket] = append(
				results[bucket],
				aws.ToString(object.Key),
			)
		}
	}

	return results
}

func assertKeysEqual(
	t *testing.T,
	actual []string,
	expected []string,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d keys, got %d: %v",
			len(expected),
			len(actual),
			actual,
		)
	}

	for index, expectedKey := range expected {
		if actual[index] != expectedKey {
			t.Fatalf(
				"expected key %d to be %q, got %q",
				index,
				expectedKey,
				actual[index],
			)
		}
	}
}

func stringID(value int) string {
	return fmt.Sprintf("%04d", value)
}
