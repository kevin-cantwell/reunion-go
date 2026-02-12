package familydata

import (
	"testing"

	reunion "github.com/kevin-cantwell/reunion-go"
)

func TestParseSchema(t *testing.T) {
	var recData []byte
	recData = append(recData, make([]byte, 6)...) // preamble

	recData = append(recData, makeTLVField(TagDisplayName, []byte("Birth"))...)
	recData = append(recData, makeTLVField(TagGEDCOMCode, []byte("BIRT"))...)
	recData = append(recData, makeTLVField(TagShortLabel, []byte("b."))...)
	recData = append(recData, makeTLVField(TagAbbreviation, []byte("Bi"))...)
	recData = append(recData, makeTLVField(TagSentenceForm, []byte("was born"))...)
	recData = append(recData, makeTLVField(TagPreposition, []byte("in"))...)

	rec := RawRecord{
		Type:   RecordTypeSchema,
		ID:     500,
		SeqNum: 10,
		Data:   recData,
	}

	ec := reunion.NewErrorCollector(0)
	def, err := ParseSchema(rec, ec)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	if def.ID != 500 {
		t.Errorf("ID = %d, want 500", def.ID)
	}
	if def.SeqNum != 10 {
		t.Errorf("SeqNum = %d, want 10", def.SeqNum)
	}
	if def.DisplayName != "Birth" {
		t.Errorf("DisplayName = %q, want %q", def.DisplayName, "Birth")
	}
	if def.GEDCOMCode != "BIRT" {
		t.Errorf("GEDCOMCode = %q, want %q", def.GEDCOMCode, "BIRT")
	}
	if def.ShortLabel != "b." {
		t.Errorf("ShortLabel = %q, want %q", def.ShortLabel, "b.")
	}
	if def.Abbreviation != "Bi" {
		t.Errorf("Abbreviation = %q, want %q", def.Abbreviation, "Bi")
	}
	if def.SentenceForm != "was born" {
		t.Errorf("SentenceForm = %q, want %q", def.SentenceForm, "was born")
	}
	if def.Preposition != "in" {
		t.Errorf("Preposition = %q, want %q", def.Preposition, "in")
	}
}

func TestParseSchema_EmptyData(t *testing.T) {
	rec := RawRecord{Type: RecordTypeSchema, ID: 1, Data: []byte{0, 0}}
	ec := reunion.NewErrorCollector(0)
	def, err := ParseSchema(rec, ec)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}
	if def.DisplayName != "" {
		t.Errorf("DisplayName = %q, want empty", def.DisplayName)
	}
}

func TestParseSchema_UnknownTagGoesToRawFields(t *testing.T) {
	var recData []byte
	recData = append(recData, make([]byte, 6)...)
	recData = append(recData, makeTLVField(0xBEEF, []byte("mystery"))...)

	rec := RawRecord{Type: RecordTypeSchema, ID: 1, Data: recData}
	ec := reunion.NewErrorCollector(0)
	def, err := ParseSchema(rec, ec)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	if len(def.RawFields) != 1 {
		t.Fatalf("RawFields count = %d, want 1", len(def.RawFields))
	}
	if def.RawFields[0].Tag != 0xBEEF {
		t.Errorf("RawFields[0].Tag = 0x%04X, want 0xBEEF", def.RawFields[0].Tag)
	}
}
