package familydata

import (
	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"

	reunion "github.com/kevin-cantwell/reunion-go"
)

// Tag constants for schema/event definition record fields.
const (
	TagDisplayName   uint16 = 0x0014
	TagGEDCOMCode    uint16 = 0x0019
	TagShortLabel    uint16 = 0x0028
	TagAbbreviation  uint16 = 0x0032
	TagAbbreviation2 uint16 = 0x0037
	TagAbbreviation3 uint16 = 0x003C
	TagSentenceForm  uint16 = 0x006E
	TagPreposition   uint16 = 0x0078
)

// ParseSchema parses a 0x20CC schema record into an EventDefinition.
func ParseSchema(rec RawRecord, ec *reunion.ErrorCollector) (*model.EventDefinition, error) {
	def := &model.EventDefinition{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 8 {
		return def, nil
	}

	data := rec.Data[8:]

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
		str := cleanString(fieldData)

		switch tag {
		case TagDisplayName:
			def.DisplayName = str
		case TagGEDCOMCode:
			def.GEDCOMCode = str
		case TagShortLabel:
			def.ShortLabel = str
		case TagAbbreviation:
			def.Abbreviation = str
		case TagAbbreviation2:
			def.Abbreviation2 = str
		case TagAbbreviation3:
			def.Abbreviation3 = str
		case TagSentenceForm:
			def.SentenceForm = str
		case TagPreposition:
			def.Preposition = str
		default:
			def.RawFields = append(def.RawFields, model.RawField{
				Tag:  tag,
				Data: fieldData,
				Size: fieldSize,
			})
		}

		pos = fieldEnd
	}

	return def, nil
}
