package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	// DefaultPresignExpiry controls how long S3 display URLs remain valid.
	DefaultPresignExpiry = 15 * time.Minute

	defaultStaticBaseURL = "https://local.wildlens.test"
)

var (
	ErrBucketRequired     = errors.New("bucket is required")
	ErrObjectPathRequired = errors.New("object path is required")
)

// MediaURLSigner generates a temporary browser-accessible URL.
type MediaURLSigner interface {
	SignGetObjectURL(
		ctx context.Context,
		bucket string,
		objectPath string,
	) (string, error)
}

// StaticURLSigner creates predictable URLs for local memory mode.
type StaticURLSigner struct {
	baseURL string
}

// NewStaticURLSigner creates a signer for local development.
func NewStaticURLSigner(baseURL string) *StaticURLSigner {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	if baseURL == "" {
		baseURL = defaultStaticBaseURL
	}

	return &StaticURLSigner{
		baseURL: baseURL,
	}
}

// SignGetObjectURL creates a predictable local display URL.
func (s *StaticURLSigner) SignGetObjectURL(
	_ context.Context,
	bucket string,
	objectPath string,
) (string, error) {
	bucket = strings.TrimSpace(bucket)
	objectPath = strings.TrimLeft(
		strings.TrimSpace(objectPath),
		"/",
	)

	if bucket == "" {
		return "", ErrBucketRequired
	}

	if objectPath == "" {
		return "", ErrObjectPathRequired
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		s.baseURL,
		url.PathEscape(bucket),
		escapeObjectPath(objectPath),
	), nil
}

// S3PresignAPI contains the S3 operation required by S3URLSigner.
type S3PresignAPI interface {
	PresignGetObject(
		ctx context.Context,
		params *s3.GetObjectInput,
		optFns ...func(*s3.PresignOptions),
	) (*v4.PresignedHTTPRequest, error)
}

// S3URLSigner creates real temporary S3 Presigned GET URLs.
type S3URLSigner struct {
	client S3PresignAPI
	expiry time.Duration
}

// NewS3URLSigner creates an S3-backed URL signer.
func NewS3URLSigner(
	client S3PresignAPI,
	expiry time.Duration,
) *S3URLSigner {
	if expiry <= 0 {
		expiry = DefaultPresignExpiry
	}

	return &S3URLSigner{
		client: client,
		expiry: expiry,
	}
}

// SignGetObjectURL creates a temporary S3 Presigned GET URL.
func (s *S3URLSigner) SignGetObjectURL(
	ctx context.Context,
	bucket string,
	objectPath string,
) (string, error) {
	bucket = strings.TrimSpace(bucket)
	objectPath = strings.TrimLeft(
		strings.TrimSpace(objectPath),
		"/",
	)

	if bucket == "" {
		return "", ErrBucketRequired
	}

	if objectPath == "" {
		return "", ErrObjectPathRequired
	}

	request, err := s.client.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectPath),
		},
		func(options *s3.PresignOptions) {
			options.Expires = s.expiry
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"presign S3 object %s/%s: %w",
			bucket,
			objectPath,
			err,
		)
	}

	if request == nil || strings.TrimSpace(request.URL) == "" {
		return "", errors.New("presigned URL is empty")
	}

	return request.URL, nil
}

func escapeObjectPath(objectPath string) string {
	segments := strings.Split(objectPath, "/")

	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}

	return strings.Join(segments, "/")
}
