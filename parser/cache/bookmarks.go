package cache

import (
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-explore/model"
)

// ParseBookmarks parses the bookmarks.cache file.
// Format: small binary file with magic "2kmb"/"2smb".
func ParseBookmarks(path string) (*model.Bookmark, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading bookmarks.cache: %w", err)
	}

	return &model.Bookmark{
		RawData: data,
		Size:    len(data),
	}, nil
}
