package familydata

import (
	"github.com/kedoco/reunion-explore/model"

	reunion "github.com/kedoco/reunion-explore"
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

	if len(rec.Data) < 6 {
		return def, nil
	}

	fields := ParseTLVFields(rec.Data)

	for _, f := range fields {
		str := cleanString(f.Data)

		switch f.Tag {
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
				Tag:  f.Tag,
				Data: f.Data,
				Size: uint16(len(f.Data) + 4),
			})
		}
	}

	return def, nil
}
