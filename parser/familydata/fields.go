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

// ExtractTLVFields parses TLV fields from record data.
// The record data (after the preamble) contains fields with:
// tag(2 LE) + data, where the field boundaries are determined by
// scanning for known tag patterns.
func ExtractTLVFields(data []byte) []TLVField {
	var fields []TLVField
	pos := 0

	for pos+4 <= len(data) {
		// Read potential field header
		// Format: some_value(2) + tag(2) or tag(2) + size/offset(2)
		// We'll try reading as: size(2) + tag(2) = 4 byte header
		fieldSize, err := binutil.U16LE(data, pos)
		if err != nil {
			break
		}
		tag, err := binutil.U16LE(data, pos+2)
		if err != nil {
			break
		}

		// Validate: size should be reasonable
		if fieldSize == 0 {
			pos += 2
			continue
		}

		fieldEnd := pos + 4 + int(fieldSize)
		if fieldEnd > len(data) {
			fieldEnd = len(data)
		}

		fieldData := data[pos+4 : fieldEnd]

		fields = append(fields, TLVField{
			Tag:    tag,
			Offset: pos,
			Data:   fieldData,
		})

		pos = fieldEnd
	}

	return fields
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
