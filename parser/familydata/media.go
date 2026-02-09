package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseMedia parses a 0x20D4 media metadata record from familydata.
func ParseMedia(rec RawRecord, ec *reunion.ErrorCollector) (*model.MediaRef, error) {
	m := &model.MediaRef{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 6 {
		return m, nil
	}

	fields := ParseTLVFields(rec.Data)

	for _, f := range fields {
		m.RawFields = append(m.RawFields, model.RawField{
			Tag:  f.Tag,
			Data: f.Data,
			Size: uint16(len(f.Data) + 4),
		})
	}

	return m, nil
}
