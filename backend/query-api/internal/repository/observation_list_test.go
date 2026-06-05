package repository

import (
	"testing"

	"github.com/jiawei-wang-dev/wildlens-cloud/backend/query-api/internal/model"
)

func TestPaginateObservationsFiltersByTagsUsingAND(t *testing.T) {
	files := []model.MediaFile{
		{
			FileID: "checksum-image-001",
			Tags:   []string{"koala", "cute", "wild"},
			TagCounts: map[string]int{
				"koala": 1,
				"cute":  1,
				"wild":  1,
			},
		},
		{
			FileID: "checksum-image-002",
			Tags:   []string{"koala"},
			TagCounts: map[string]int{
				"koala": 1,
			},
		},
	}

	page, err := paginateObservations(
		files,
		ObservationListOptions{
			Limit: 10,
			Tags:  []string{"koala", "cute"},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}

	if page.Items[0].FileID != "checksum-image-001" {
		t.Fatalf("unexpected file ID: %s", page.Items[0].FileID)
	}
}

func TestPaginateObservationsMatchesTagsFallback(t *testing.T) {
	files := []model.MediaFile{
		{
			FileID: "checksum-image-001",
			Tags:   []string{"Koala"},
		},
	}

	page, err := paginateObservations(
		files,
		ObservationListOptions{
			Limit: 10,
			Tags:  []string{" koala "},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
}

func TestPaginateObservationsPaginatesAfterTagFilter(t *testing.T) {
	files := []model.MediaFile{
		{
			FileID: "checksum-image-001",
			TagCounts: map[string]int{
				"wild": 1,
			},
		},
		{
			FileID: "checksum-image-002",
			TagCounts: map[string]int{
				"wild": 1,
			},
		},
		{
			FileID: "checksum-image-003",
			TagCounts: map[string]int{
				"domestic": 1,
			},
		},
	}

	firstPage, err := paginateObservations(
		files,
		ObservationListOptions{
			Limit: 1,
			Tags:  []string{"wild"},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(firstPage.Items) != 1 {
		t.Fatalf("expected 1 first-page item, got %d", len(firstPage.Items))
	}

	if !firstPage.HasMore {
		t.Fatal("expected first page to have more results")
	}

	secondPage, err := paginateObservations(
		files,
		ObservationListOptions{
			Limit:     1,
			NextToken: firstPage.NextToken,
			Tags:      []string{"wild"},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(secondPage.Items) != 1 {
		t.Fatalf("expected 1 second-page item, got %d", len(secondPage.Items))
	}

	if secondPage.Items[0].FileID != "checksum-image-002" {
		t.Fatalf(
			"unexpected second-page file ID: %s",
			secondPage.Items[0].FileID,
		)
	}

	if secondPage.HasMore {
		t.Fatal("expected second page to have no more results")
	}
}
