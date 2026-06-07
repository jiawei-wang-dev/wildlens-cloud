package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("temporary image inference is not configured")

// ImageResult contains stateless inference output for one temporary image.
type ImageResult struct {
	Tags           []string       `json:"tags"`
	TagCounts      map[string]int `json:"tag_counts"`
	PrimarySpecies string         `json:"primary_species"`
	ModelVersion   string         `json:"model_version"`
}

// ImageClient runs stateless inference for temporary query images.
type ImageClient interface {
	InferImage(
		ctx context.Context,
		filename string,
		contentType string,
		data []byte,
	) (ImageResult, error)
}

// NoopClient avoids external calls when TEMP_QUERY_INFER_URL is not configured.
type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (c *NoopClient) InferImage(
	_ context.Context,
	_ string,
	_ string,
	_ []byte,
) (ImageResult, error) {
	return ImageResult{}, ErrNotConfigured
}

// HTTPClient calls a stateless media inference endpoint.
type HTTPClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewHTTPClient(
	endpoint string,
	httpClient *http.Client,
) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &HTTPClient{
		endpoint:   strings.TrimSpace(endpoint),
		httpClient: httpClient,
	}
}

func (c *HTTPClient) InferImage(
	ctx context.Context,
	filename string,
	contentType string,
	data []byte,
) (ImageResult, error) {
	if c.endpoint == "" {
		return ImageResult{}, ErrNotConfigured
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return ImageResult{}, fmt.Errorf("create inference multipart file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return ImageResult{}, fmt.Errorf("write inference multipart file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return ImageResult{}, fmt.Errorf("close inference multipart body: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		body,
	)
	if err != nil {
		return ImageResult{}, fmt.Errorf("create inference request: %w", err)
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	if contentType != "" {
		request.Header.Set("X-WildLens-Content-Type", contentType)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return ImageResult{}, fmt.Errorf("call inference service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return ImageResult{}, fmt.Errorf(
			"inference service returned status %d",
			response.StatusCode,
		)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ImageResult{}, fmt.Errorf("read inference response: %w", err)
	}

	var result ImageResult

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return ImageResult{}, fmt.Errorf("decode inference response: %w", err)
	}

	return result, nil
}
