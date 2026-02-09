package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
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

	for _, f := range fields {
		str := cleanString(f.Data)
		if s.Title == "" && len(str) > 2 {
			s.Title = str
		}

		s.RawFields = append(s.RawFields, model.RawField{
			Tag:  f.Tag,
			Data: f.Data,
			Size: uint16(len(f.Data) + 4),
		})
	}

	return s, nil
}
