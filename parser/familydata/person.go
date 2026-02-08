package familydata

import (
	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"

	reunion "github.com/kevin-cantwell/reunion-go"
)

// Tag constants for person record fields.
const (
	TagGivenName uint16 = 0x001E
	TagSurname1  uint16 = 0x000C
	TagSurname2  uint16 = 0x0023
	TagSexFlags  uint16 = 0x001B
)

// ParsePerson parses a 0x20C4 person record into a Person model.
func ParsePerson(rec RawRecord, ec *reunion.ErrorCollector) (*model.Person, error) {
	p := &model.Person{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 8 {
		return p, nil
	}

	// Skip the first 8 bytes (timestamp + flags)
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

		switch tag {
		case TagGivenName:
			p.GivenName = cleanString(fieldData)
		case TagSurname1, TagSurname2:
			if p.Surname == "" {
				p.Surname = cleanString(fieldData)
			}
		case TagSexFlags:
			if len(fieldData) >= 1 {
				switch fieldData[0] {
				case 1:
					p.Sex = model.SexMale
				case 2:
					p.Sex = model.SexFemale
				}
			}
		default:
			// Check for events containing [[pt:NNN]]
			placeRefs := ExtractPlaceRefs(fieldData)
			if len(placeRefs) > 0 || isEventTag(tag) {
				evt := model.PersonEvent{
					Tag:       tag,
					PlaceRefs: placeRefs,
					RawData:   fieldData,
				}
				p.Events = append(p.Events, evt)
			} else {
				p.RawFields = append(p.RawFields, model.RawField{
					Tag:  tag,
					Data: fieldData,
					Size: fieldSize,
				})
			}
		}

		pos = fieldEnd
	}

	return p, nil
}

func cleanString(data []byte) string {
	// Find printable string content
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] != 0 {
			return string(data[:i+1])
		}
	}
	return ""
}

func isEventTag(tag uint16) bool {
	// Event tags typically use codes >= 0x002B in the person record
	// These are the field tags for birth, death, and other events
	return tag >= 0x002B && tag <= 0x00FF
}
