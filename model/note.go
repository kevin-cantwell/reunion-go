package model

// Note represents a note, either from a .note file or an inline familydata record.
type Note struct {
	ID          uint32       `json:"id"`
	SeqNum      uint16       `json:"seq_num,omitempty"`
	PersonID    int          `json:"person_id,omitempty"`
	EventTag    int          `json:"event_tag,omitempty"`
	SourceID    int          `json:"source_id,omitempty"`
	Filename    string       `json:"filename,omitempty"`
	RawText     string       `json:"raw_text,omitempty"`
	Markup      []MarkupNode `json:"markup,omitempty"`
	DisplayText string       `json:"display_text,omitempty"`
}

// NoteRef is a reference from a person or family to a note.
type NoteRef struct {
	NoteID   uint32 `json:"note_id"`
	EventTag uint16 `json:"event_tag,omitempty"`
	SchemaID uint16 `json:"schema_id,omitempty"`
}
