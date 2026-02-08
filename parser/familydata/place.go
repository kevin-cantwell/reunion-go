package familydata

import (
	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParsePlace parses a 0x20D8 place record from familydata.
func ParsePlace(rec RawRecord, ec *reunion.ErrorCollector) (*model.Place, error) {
	p := &model.Place{
		ID: rec.ID,
	}

	if len(rec.Data) > 8 {
		// Extract place name from record data
		data := rec.Data[8:]
		p.Name = ExtractString(data)
	}

	return p, nil
}
