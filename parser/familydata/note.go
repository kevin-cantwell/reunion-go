package familydata

import (
	"bytes"

	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/model"
	"github.com/kevin-cantwell/reunion-go/parser/notes"
)

// ParseNote parses a 0x2104 inline note record from familydata.
func ParseNote(rec RawRecord, ec *reunion.ErrorCollector) (*model.Note, error) {
	n := &model.Note{
		ID:     rec.ID,
		SeqNum: rec.SeqNum,
	}

	if len(rec.Data) > 8 {
		data := rec.Data[8:]
		text := extractNoteText(data)
		n.RawText = text
		if text != "" {
			n.Markup = notes.ParseMarkup(text)
			n.DisplayText = model.PlainText(n.Markup)
		}
	}

	return n, nil
}

// openTagBytes is the UTF-8 encoding of « (U+00AB).
var openTagBytes = []byte{0xC2, 0xAB}

func extractNoteText(data []byte) string {
	// Inline notes have a binary preamble followed by text content.
	// The text typically starts with «ff=1» markup.
	// Look for the first occurrence of the « character (0xC2 0xAB).
	idx := bytes.Index(data, openTagBytes)
	if idx >= 0 {
		return string(data[idx:])
	}

	// Fallback: find the first run of printable text (at least 5 chars)
	for i := 0; i < len(data); i++ {
		if data[i] >= 0x20 && data[i] < 0x7F {
			printable := 0
			for j := i; j < len(data) && j < i+10; j++ {
				if data[j] >= 0x20 && data[j] < 0x7F {
					printable++
				}
			}
			if printable >= 5 {
				return string(data[i:])
			}
		}
	}

	return ""
}
