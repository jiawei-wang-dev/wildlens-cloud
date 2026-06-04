package repository

import "strings"

// newFileIDSet trims file IDs, removes empty values and removes duplicates.
func newFileIDSet(fileIDs []string) map[string]struct{} {
	results := make(map[string]struct{})

	for _, fileID := range fileIDs {
		fileID = strings.TrimSpace(fileID)

		if fileID == "" {
			continue
		}

		results[fileID] = struct{}{}
	}

	return results
}
