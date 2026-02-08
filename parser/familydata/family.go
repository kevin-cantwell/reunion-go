package familydata

import (
	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"

	reunion "github.com/kevin-cantwell/reunion-go"
)

// Tag constants for family record fields.
const (
	TagPartner1  uint16 = 0x0050
	TagPartner2  uint16 = 0x0051
	TagChild1    uint16 = 0x00FA
	TagChild2    uint16 = 0x00FB
	TagChild3    uint16 = 0x00FC
	TagMarriage  uint16 = 0x005F
)

// ParseFamily parses a 0x20C8 family record into a Family model.
func ParseFamily(rec RawRecord, ec *reunion.ErrorCollector) (*model.Family, error) {
	f := &model.Family{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) < 8 {
		return f, nil
	}

	// Skip timestamp bytes
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

		switch tag {
		case TagPartner1:
			if len(fieldData) >= 4 {
				id, _ := binutil.U32LE(fieldData, 0)
				f.Partner1 = id
			} else if len(fieldData) >= 2 {
				id, _ := binutil.U16LE(fieldData, 0)
				f.Partner1 = uint32(id)
			}
		case TagPartner2:
			if len(fieldData) >= 4 {
				id, _ := binutil.U32LE(fieldData, 0)
				f.Partner2 = id
			} else if len(fieldData) >= 2 {
				id, _ := binutil.U16LE(fieldData, 0)
				f.Partner2 = uint32(id)
			}
		case TagChild1, TagChild2, TagChild3:
			if len(fieldData) >= 4 {
				id, _ := binutil.U32LE(fieldData, 0)
				if id > 0 {
					f.Children = append(f.Children, id)
				}
			} else if len(fieldData) >= 2 {
				id, _ := binutil.U16LE(fieldData, 0)
				if id > 0 {
					f.Children = append(f.Children, uint32(id))
				}
			}
		default:
			placeRefs := ExtractPlaceRefs(fieldData)
			if len(placeRefs) > 0 || tag == TagMarriage {
				evt := model.FamilyEvent{
					Tag:       tag,
					PlaceRefs: placeRefs,
					RawData:   fieldData,
				}
				f.Events = append(f.Events, evt)
			} else {
				f.RawFields = append(f.RawFields, model.RawField{
					Tag:  tag,
					Data: fieldData,
					Size: fieldSize,
				})
			}
		}

		pos = fieldEnd
	}

	return f, nil
}
