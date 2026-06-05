package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3PresignClient struct {
	input  *s3.GetObjectInput
	expiry time.Duration
	url    string
	err    error
}

func (f *fakeS3PresignClient) PresignGetObject(
	_ context.Context,
	params *s3.GetObjectInput,
	optFns ...func(*s3.PresignOptions),
) (*v4.PresignedHTTPRequest, error) {
	if f.err != nil {
		return nil, f.err
	}

	options := s3.PresignOptions{}

	for _, option := range optFns {
		option(&options)
	}

	f.input = params
	f.expiry = options.Expires

	return &v4.PresignedHTTPRequest{
		URL: f.url,
	}, nil
}

func TestStaticURLSignerBuildsPredictableURL(t *testing.T) {
	signer := NewStaticURLSigner(
		"https://local.wildlens.test",
	)

	signedURL, err := signer.SignGetObjectURL(
		context.Background(),
		"wildlens-media",
		"media/thumbnails/koala image.jpg",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "https://local.wildlens.test/" +
		"wildlens-media/media/thumbnails/koala%20image.jpg"

	if signedURL != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			signedURL,
		)
	}
}

func TestS3URLSignerPresignsGetObject(t *testing.T) {
	client := &fakeS3PresignClient{
		url: "https://signed.example/koala.jpg",
	}

	signer := NewS3URLSigner(
		client,
		5*time.Minute,
	)

	signedURL, err := signer.SignGetObjectURL(
		context.Background(),
		"wildlens-media",
		"media/originals/koala.jpg",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if signedURL != "https://signed.example/koala.jpg" {
		t.Fatalf("unexpected signed URL: %s", signedURL)
	}

	if aws.ToString(client.input.Bucket) != "wildlens-media" {
		t.Fatalf(
			"unexpected bucket: %s",
			aws.ToString(client.input.Bucket),
		)
	}

	if aws.ToString(client.input.Key) != "media/originals/koala.jpg" {
		t.Fatalf(
			"unexpected object key: %s",
			aws.ToString(client.input.Key),
		)
	}

	if client.expiry != 5*time.Minute {
		t.Fatalf(
			"unexpected expiry: %s",
			client.expiry,
		)
	}
}

func TestS3URLSignerReturnsPresignError(t *testing.T) {
	client := &fakeS3PresignClient{
		err: errors.New("presign failed"),
	}

	signer := NewS3URLSigner(
		client,
		DefaultPresignExpiry,
	)

	_, err := signer.SignGetObjectURL(
		context.Background(),
		"wildlens-media",
		"media/originals/koala.jpg",
	)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
