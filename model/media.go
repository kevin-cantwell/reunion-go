package model

// MediaRef represents a media/attachment record from the familydata.
type MediaRef struct {
	ID        uint32     `json:"id"`
	SeqNum    uint16     `json:"seq_num"`
	RawFields []RawField `json:"raw_fields,omitempty"`
}
