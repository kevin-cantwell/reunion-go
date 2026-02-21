package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-explore"
	"github.com/kevin-cantwell/reunion-explore/model"
)

// ParseSource parses a 0x20D0 source record from familydata.
func ParseSource(rec RawRecord, ec *reunion.ErrorCollector) (*model.Source, error) {
	s := &model.Source{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 6 {
		return s, nil
	}

	fields := ParseTLVFields(rec.Data)

	// First pass: look for tag 0x0014 (display name) for the title
	for _, f := range fields {
		if f.Tag == TagDisplayName {
			str := cleanString(f.Data)
			if len(str) > 0 {
				s.Title = str
			}
		}
		s.RawFields = append(s.RawFields, model.RawField{
			Tag:  f.Tag,
			Data: f.Data,
			Size: uint16(len(f.Data) + 4),
		})
	}

	// Fallback: if no 0x0014 tag, use first non-empty string
	if s.Title == "" {
		for _, f := range fields {
			str := cleanString(f.Data)
			if len(str) > 2 {
				s.Title = str
				break
			}
		}
	}

	return s, nil
}
