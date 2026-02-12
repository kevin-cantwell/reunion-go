package model

// SourceCitation represents a reference to a source record,
// optionally with detail text (e.g., "page 274").
type SourceCitation struct {
	SourceID uint32 `json:"source_id"`
	Detail   string `json:"detail,omitempty"`
}
