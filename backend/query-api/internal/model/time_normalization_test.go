package model

import "testing"

func TestNormalizeTimestampUTC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "RFC3339 Z",
			input:    "2026-06-07T07:56:13Z",
			expected: "2026-06-07T07:56:13Z",
		},
		{
			name:     "RFC3339 microseconds Z",
			input:    "2026-06-07T07:56:13.211614Z",
			expected: "2026-06-07T07:56:13Z",
		},
		{
			name:     "explicit timezone",
			input:    "2026-06-07T17:56:13+10:00",
			expected: "2026-06-07T07:56:13Z",
		},
		{
			name:     "historical UTC without timezone",
			input:    "2026-06-07T07:56:13",
			expected: "2026-06-07T07:56:13Z",
		},
		{
			name:     "historical UTC microseconds without timezone",
			input:    "2026-06-07T07:56:13.211614",
			expected: "2026-06-07T07:56:13Z",
		},
		{
			name:     "empty value",
			input:    "",
			expected: "",
		},
		{
			name:     "unparseable value",
			input:    "legacy-time",
			expected: "legacy-time",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NormalizeTimestampUTC(test.input)
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestNormalizeMediaFileTimes(t *testing.T) {
	file := MediaFile{
		CreatedAt: "2026-06-07T07:54:25.982264",
		UpdatedAt: "2026-06-07T17:54:25+10:00",
	}

	NormalizeMediaFileTimes(&file)

	if file.CreatedAt != "2026-06-07T07:54:25Z" {
		t.Fatalf("unexpected created_at: %s", file.CreatedAt)
	}

	if file.UpdatedAt != "2026-06-07T07:54:25Z" {
		t.Fatalf("unexpected updated_at: %s", file.UpdatedAt)
	}
}
