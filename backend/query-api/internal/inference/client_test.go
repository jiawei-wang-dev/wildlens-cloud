package inference

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return f(request)
}

func TestHTTPClientInferImageSendsMultipartFile(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(
			func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Fatalf("expected POST, got %s", r.Method)
				}

				if !strings.HasPrefix(
					r.Header.Get("Content-Type"),
					"multipart/form-data;",
				) {
					t.Fatalf(
						"expected multipart content type, got %q",
						r.Header.Get("Content-Type"),
					)
				}

				if r.Header.Get("X-WildLens-Content-Type") != "image/jpeg" {
					t.Fatalf(
						"expected original content type header, got %q",
						r.Header.Get("X-WildLens-Content-Type"),
					)
				}

				file, header, err := r.FormFile("file")
				if err != nil {
					t.Fatalf("expected multipart file field: %v", err)
				}
				defer file.Close()

				if header.Filename != "query.jpg" {
					t.Fatalf(
						"expected filename query.jpg, got %q",
						header.Filename,
					)
				}

				responseBody, err := json.Marshal(ImageResult{
					Tags: []string{"koala"},
					TagCounts: map[string]int{
						"koala": 1,
					},
					PrimarySpecies: "koala",
					ModelVersion:   "provided-aussie-ecolense-v1",
				})
				if err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(
						strings.NewReader(string(responseBody)),
					),
				}, nil
			},
		),
	}

	client := NewHTTPClient("https://inference.test/query", httpClient)

	result, err := client.InferImage(
		context.Background(),
		"query.jpg",
		"image/jpeg",
		[]byte("image bytes"),
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Tags) != 1 || result.Tags[0] != "koala" {
		t.Fatalf("unexpected tags: %#v", result.Tags)
	}

	if result.TagCounts["koala"] != 1 {
		t.Fatalf("unexpected tag count: %#v", result.TagCounts)
	}
}

func TestHTTPClientInferImageReturnsStatusError(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(
			func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       io.NopCloser(strings.NewReader("failed")),
				}, nil
			},
		),
	}

	client := NewHTTPClient("https://inference.test/query", httpClient)

	_, err := client.InferImage(
		context.Background(),
		"query.jpg",
		"image/jpeg",
		[]byte("image bytes"),
	)
	if err == nil {
		t.Fatal("expected inference status error")
	}
}
