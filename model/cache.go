package model

// FirstNameEntry represents an entry from fmnames.cache.
type FirstNameEntry struct {
	Name     string `json:"name"`
	Meta     []byte `json:"-"`
	Phonetic string `json:"phonetic,omitempty"`
}

// SurnameEntry represents an entry from surnames.cache.
type SurnameEntry struct {
	Surname   string `json:"surname"`
	GivenName string `json:"given_name,omitempty"`
	RawEntry  string `json:"raw_entry,omitempty"`
}

// SearchName represents an entry from shNames.cache.
type SearchName struct {
	Name string `json:"name"`
}

// TimestampEntry represents a 20-byte record from timestamps.cache.
type TimestampEntry struct {
	Data []byte `json:"-"`
	Hex  string `json:"hex"`
}

// Bookmark represents the bookmarks.cache content.
type Bookmark struct {
	RawData []byte `json:"-"`
	Size    int    `json:"size"`
}

// ColorTag represents a color tag from colortags.cache.
type ColorTag struct {
	Data []byte `json:"-"`
}

// Association represents an entry from associations.cache.
type Association struct {
	Data []byte `json:"-"`
}

// GlobalRecordEntry represents the globalRecords.cache content.
type GlobalRecordEntry struct {
	RawData []byte `json:"-"`
	Hex     string `json:"hex"`
}
