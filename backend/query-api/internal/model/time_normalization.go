package model

import (
	"strings"
	"time"
)

// NormalizeTimestampUTC returns a UTC RFC3339 timestamp when the value is parseable.
// Unparseable historical values are preserved so one bad record never breaks an API response.
func NormalizeTimestampUTC(value string) string {
	parsed, ok := ParseTimestampUTC(value)
	if !ok {
		return value
	}

	return parsed.UTC().Format(time.RFC3339)
}

func NormalizeMediaFileTimes(file *MediaFile) {
	if file == nil {
		return
	}

	file.CreatedAt = NormalizeTimestampUTC(file.CreatedAt)
	file.UpdatedAt = NormalizeTimestampUTC(file.UpdatedAt)
}

func NormalizeMediaFilesTimes(files []MediaFile) {
	for index := range files {
		NormalizeMediaFileTimes(&files[index])
	}
}

func ParseTimestampUTC(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)

	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), true
		}
	}

	return time.Time{}, false
}
