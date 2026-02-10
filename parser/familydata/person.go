package familydata

import (
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

	if len(rec.Data) < 6 {
		return p, nil
	}

	fields := ParseTLVFields(rec.Data)

	for _, f := range fields {
		switch f.Tag {
		case TagGivenName:
			p.GivenName = cleanString(f.Data)
		case TagSurname1, TagSurname2:
			if p.Surname == "" {
				p.Surname = cleanString(f.Data)
			}
		case TagSexFlags:
			if len(f.Data) >= 1 {
				switch f.Data[0] {
				case 1:
					p.Sex = model.SexMale
				case 2:
					p.Sex = model.SexFemale
				}
			}
		default:
			if isEventTag(f.Tag) {
				evt := model.PersonEvent{
					Tag:       f.Tag,
					PlaceRefs: ExtractPlaceRefs(f.Data),
					RawData:   f.Data,
					SchemaID:  ParseEventField(f.Data),
					Date:      ExtractDate(f.Data),
				}
				p.Events = append(p.Events, evt)

				// For note-type events (tag < 0x03E8), extract the note reference
				if f.Tag < 0x03E8 {
					noteID := ExtractNoteRef(f.Data)
					if noteID > 0 {
						p.NoteRefs = append(p.NoteRefs, model.NoteRef{
							NoteID:   noteID,
							EventTag: f.Tag,
							SchemaID: evt.SchemaID,
						})
					}
				}
			} else {
				p.RawFields = append(p.RawFields, model.RawField{
					Tag:  f.Tag,
					Data: f.Data,
					Size: uint16(len(f.Data) + 4),
				})
			}
		}
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
	// Event tags use codes >= 0x100 in the person record
	return tag >= 0x0100
}
