package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseSource parses a 0x20D0 source record from familydata.
func ParseSource(rec RawRecord, ec *reunion.ErrorCollector) (*model.Source, error) {
	s := &model.Source{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 8 {
		return s, nil
	}

	data := rec.Data[8:]

	// Parse TLV fields looking for a title string
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

		// Try to find title in early fields
		str := cleanString(fieldData)
		if s.Title == "" && len(str) > 2 {
			s.Title = str
		}

		s.RawFields = append(s.RawFields, model.RawField{
			Tag:  tag,
			Data: fieldData,
			Size: fieldSize,
		})

		pos = fieldEnd
	}

	return s, nil
}
