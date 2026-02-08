package model

// Source represents a source record from the familydata.
type Source struct {
	ID        uint32     `json:"id"`
	SeqNum    uint16     `json:"seq_num"`
	Title     string     `json:"title,omitempty"`
	RawFields []RawField `json:"raw_fields,omitempty"`
}
