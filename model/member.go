package model

// Member represents a sync member directory within the bundle.
type Member struct {
	Name       string         `json:"name"`
	DirPath    string         `json:"dir_path,omitempty"`
	NoteFiles  []string       `json:"note_files,omitempty"`
	HasChanges bool           `json:"has_changes"`
	HasMedia   bool           `json:"has_media"`
	Changes    []ChangeRecord `json:"changes,omitempty"`
}

// ChangeRecord represents a single entry from a .changes file.
type ChangeRecord struct {
	Offset int    `json:"offset"`
	Size   int    `json:"size"`
	Data   []byte `json:"-"`
}
