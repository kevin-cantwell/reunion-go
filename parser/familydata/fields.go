package familydata

import (
	"regexp"
	"strconv"

	"github.com/kevin-cantwell/reunion-go/internal/binutil"
)

// placeRefPattern matches [[pt:NNN]] place references in event data.
var placeRefPattern = regexp.MustCompile(`\[\[pt:(\d+)\]\]`)

// TLVField represents a tag-length-value field extracted from record data.
type TLVField struct {
	Tag    uint16
	Offset int // offset within record data
	Data   []byte
}

// ParseTLVFields parses TLV fields from raw record data.
// Record data starts with a 6-byte preamble (4-byte timestamp + 2-byte repeated size),
// followed by fields in the format: total_length(u16 LE) + tag(u16 LE) + data(total_length - 4).
func ParseTLVFields(data []byte) []TLVField {
	if len(data) < 6 {
		return nil
	}

	// Skip preamble: 4-byte timestamp + 2-byte repeated size
	return parseTLVFieldsFrom(data[6:])
}

// parseTLVFieldsFrom parses TLV fields starting at an arbitrary position.
func parseTLVFieldsFrom(data []byte) []TLVField {
	var fields []TLVField
	pos := 0

	for pos+4 <= len(data) {
		totalLen, err := binutil.U16LE(data, pos)
		if err != nil {
			break
		}
		tag, err := binutil.U16LE(data, pos+2)
		if err != nil {
			break
		}

		// total_length includes the 4-byte header itself
		if totalLen < 4 {
			break
		}

		dataLen := int(totalLen) - 4
		fieldEnd := pos + int(totalLen)
		if fieldEnd > len(data) {
			fieldEnd = len(data)
			dataLen = fieldEnd - pos - 4
			if dataLen < 0 {
				break
			}
		}

		fieldData := data[pos+4 : pos+4+dataLen]

		fields = append(fields, TLVField{
			Tag:    tag,
			Offset: pos,
			Data:   fieldData,
		})

		pos = fieldEnd
	}

	return fields
}

// ParseEventField extracts the schema definition ID from event sub-data.
// Event fields (tags >= 0x100) contain nested data. After a 4-byte sub-header,
// the schema definition ID is at offset 12 within the sub-data as u16 LE.
func ParseEventField(fieldData []byte) (schemaID uint16) {
	// Sub-data starts after the 4-byte sub-header
	// Schema ID is at offset 12 within the sub-data (i.e., byte 16 from start of fieldData)
	// But per the plan: "offset 12 in sub-data (after the 4-byte sub-header)"
	// So: 4 (sub-header) + 12 = offset 16 from start of fieldData
	if len(fieldData) >= 18 {
		id, err := binutil.U16LE(fieldData, 16)
		if err == nil {
			schemaID = id
		}
	}
	return
}

// ExtractPlaceRefs finds all [[pt:NNN]] references in data and returns the place IDs.
func ExtractPlaceRefs(data []byte) []int {
	matches := placeRefPattern.FindAllSubmatch(data, -1)
	var refs []int
	for _, m := range matches {
		if len(m) >= 2 {
			n, err := strconv.Atoi(string(m[1]))
			if err == nil {
				refs = append(refs, n)
			}
		}
	}
	return refs
}

// ExtractString finds a printable string in the given data.
func ExtractString(data []byte) string {
	// Find the first run of printable bytes
	start := -1
	for i, b := range data {
		if b >= 0x20 && b < 0x7F {
			if start == -1 {
				start = i
			}
		} else if start != -1 {
			return string(data[start:i])
		}
	}
	if start != -1 {
		return string(data[start:])
	}
	return ""
}
