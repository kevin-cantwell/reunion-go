package familydata

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"

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

// ExtractDate decodes a date from event sub-data.
// Event data has an 18-byte fixed header, followed by sub-TLV fields.
// The first sub-TLV (at offset 18) contains the date when its total length == 8.
//
// Year and quarter are encoded in bytes[24:25] as u16 LE:
//
//	totalQ = (year + 8000) * 4 + quarter
//	quarter: 0=Q1(Jan-Mar), 1=Q2(Apr-Jun), 2=Q3(Jul-Sep), 3=Q4(Oct-Dec)
//
// Month within the quarter is encoded in the event header at bytes[2:3]:
//
//	u16LE(bytes[2:3]) % 3 maps via [0→1st, 1→3rd, 2→2nd] to month offset
//
// Day and qualifier are in bytes[22:23]:
//
//	byte[22]: precision flags (0x00=normal, 0xA0=approximate, 0x40=after)
//	byte[23]: day (bits 4-0, 0=unknown); bit 7 = day-present flag; bits 7-6 = 11 means "before"
func ExtractDate(fieldData []byte) string {
	if len(fieldData) < 26 {
		return ""
	}
	// Sub-TLV at offset 18: date sub-TLV has exactly length 8 (4-byte header + 4-byte date)
	subLen, err := binutil.U16LE(fieldData, 18)
	if err != nil || subLen != 8 {
		return ""
	}

	precFlags := fieldData[22]
	dayByte := fieldData[23]
	monthYearLo := fieldData[24]
	monthYearHi := fieldData[25]

	totalQ := int(monthYearHi)<<8 | int(monthYearLo)
	year := totalQ/4 - 8000
	quarter := totalQ % 4 // 0=Q1(Jan-Mar), 1=Q2(Apr-Jun), 2=Q3(Jul-Sep), 3=Q4(Oct-Dec)

	// Month within quarter is encoded in the event header u16LE at bytes[2:3].
	// The value mod 3 maps: 0→1st month, 1→3rd month, 2→2nd month of the quarter.
	hdrVal, _ := binutil.U16LE(fieldData, 2)
	monthOffsetTable := [3]int{0, 2, 1} // u16%3 → month offset (0-indexed within quarter)
	monthInQuarter := monthOffsetTable[int(hdrVal)%3]
	month := quarter*3 + monthInQuarter + 1

	isBefore := dayByte&0xC0 == 0xC0 // bits 7-6 both set = "before"
	day := int(dayByte & 0x1F)

	if year < 1 || year > 9999 {
		return ""
	}

	monthNames := [13]string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	prefix := ""
	switch {
	case precFlags == 0xA0 || precFlags == 0xE0:
		prefix = "about "
	case isBefore:
		prefix = "before "
	case precFlags == 0x40:
		prefix = "after "
	}

	if day > 0 && day <= 31 {
		return fmt.Sprintf("%s%d %s %d", prefix, day, monthNames[month], year)
	}
	return fmt.Sprintf("%s%s %d", prefix, monthNames[month], year)
}

// ExtractNoteRef extracts a note record ID from event sub-data.
// Note-type events (tag < 0x03E8) may contain a sub-TLV at offset 18
// with tag 0x0000 whose 4-byte data is the referenced note's record ID.
// Returns 0 if no note reference is found.
func ExtractNoteRef(fieldData []byte) uint32 {
	pos := 18
	for pos+4 <= len(fieldData) {
		subLen, err := binutil.U16LE(fieldData, pos)
		if err != nil || subLen < 4 {
			break
		}
		subTag, err := binutil.U16LE(fieldData, pos+2)
		if err != nil {
			break
		}
		if pos+int(subLen) > len(fieldData) {
			break
		}
		if subTag == 0x0000 && subLen == 8 {
			id, err := binutil.U32LE(fieldData, pos+4)
			if err == nil && id > 0 {
				return id
			}
		}
		pos += int(subLen)
	}
	return 0
}

// ExtractString finds the first run of printable characters (ASCII and valid
// UTF-8) in data. It uses unicode/utf8.DecodeRune for correct multi-byte
// handling, including rejection of overlong encodings and surrogates.
func ExtractString(data []byte) string {
	start := -1
	i := 0
	for i < len(data) {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size <= 1 {
			// Invalid UTF-8 byte — string boundary
			if start != -1 {
				return string(data[start:i])
			}
			i++
			continue
		}
		if r >= 0x20 && r != 0x7F && unicode.IsPrint(r) {
			if start == -1 {
				start = i
			}
			i += size
		} else {
			// Control char or non-printable — string boundary
			if start != -1 {
				return string(data[start:i])
			}
			i += size
		}
	}
	if start != -1 {
		return string(data[start:])
	}
	return ""
}
