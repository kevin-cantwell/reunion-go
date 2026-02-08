package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseMedia parses a 0x20D4 media metadata record from familydata.
func ParseMedia(rec RawRecord, ec *reunion.ErrorCollector) (*model.MediaRef, error) {
	m := &model.MediaRef{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 8 {
		return m, nil
	}

	data := rec.Data[8:]

	// Parse TLV fields
	pos := 0
	for pos+4 <= len(data) {
		fieldSize, err := binutil.U16LE(data, pos)
		if err != nil || fieldSize == 0 {
			pos += 2
			continue
		}
		tag, err := binutil.U16LE(data, pos+2)
		if err != nil {
			break
		}

		fieldEnd := pos + 4 + int(fieldSize)
		if fieldEnd > len(data) {
			fieldEnd = len(data)
		}
		fieldData := data[pos+4 : fieldEnd]

		m.RawFields = append(m.RawFields, model.RawField{
			Tag:  tag,
			Data: fieldData,
			Size: fieldSize,
		})

		pos = fieldEnd
	}

	return m, nil
}
