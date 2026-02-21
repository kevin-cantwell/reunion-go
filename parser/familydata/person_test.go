package familydata

import (
	"encoding/binary"
	"testing"

	reunion "github.com/kevin-cantwell/reunion-explore"
	"github.com/kevin-cantwell/reunion-explore/model"
)

func TestParsePerson(t *testing.T) {
	// Build TLV record data: 6-byte preamble + fields
	var recData []byte
	recData = append(recData, make([]byte, 6)...) // preamble

	// Given name field (tag 0x001E)
	recData = append(recData, makeTLVField(TagGivenName, []byte("Alice"))...)

	// Surname field (tag 0x000C)
	recData = append(recData, makeTLVField(TagSurname1, []byte("Smith"))...)

	// Sex field (tag 0x001B): 1 = male
	recData = append(recData, makeTLVField(TagSexFlags, []byte{0x02})...)

	// Event field (tag 0x0100, >= 0x0100 means event)
	// Build minimal event data (needs at least 18 bytes for ExtractDate to try)
	eventData := make([]byte, 20)
	recData = append(recData, makeTLVField(0x0100, eventData)...)

	rec := RawRecord{
		Type:   RecordTypePerson,
		ID:     42,
		SeqNum: 7,
		Data:   recData,
	}

	ec := reunion.NewErrorCollector(0)
	person, err := ParsePerson(rec, ec)
	if err != nil {
		t.Fatalf("ParsePerson() error = %v", err)
	}

	if person.ID != 42 {
		t.Errorf("ID = %d, want 42", person.ID)
	}
	if person.SeqNum != 7 {
		t.Errorf("SeqNum = %d, want 7", person.SeqNum)
	}
	if person.GivenName != "Alice" {
		t.Errorf("GivenName = %q, want %q", person.GivenName, "Alice")
	}
	if person.Surname != "Smith" {
		t.Errorf("Surname = %q, want %q", person.Surname, "Smith")
	}
	if person.Sex != model.SexFemale {
		t.Errorf("Sex = %d, want SexFemale (%d)", person.Sex, model.SexFemale)
	}
	if len(person.Events) != 1 {
		t.Errorf("Events count = %d, want 1", len(person.Events))
	} else if person.Events[0].Tag != 0x0100 {
		t.Errorf("Events[0].Tag = 0x%04X, want 0x0100", person.Events[0].Tag)
	}
}

func TestParsePerson_Male(t *testing.T) {
	var recData []byte
	recData = append(recData, make([]byte, 6)...)
	recData = append(recData, makeTLVField(TagSexFlags, []byte{0x01})...)

	rec := RawRecord{Type: RecordTypePerson, ID: 1, Data: recData}
	ec := reunion.NewErrorCollector(0)
	person, err := ParsePerson(rec, ec)
	if err != nil {
		t.Fatalf("ParsePerson() error = %v", err)
	}
	if person.Sex != model.SexMale {
		t.Errorf("Sex = %d, want SexMale (%d)", person.Sex, model.SexMale)
	}
}

func TestParsePerson_EmptyData(t *testing.T) {
	rec := RawRecord{Type: RecordTypePerson, ID: 1, Data: []byte{0, 0}}
	ec := reunion.NewErrorCollector(0)
	person, err := ParsePerson(rec, ec)
	if err != nil {
		t.Fatalf("ParsePerson() error = %v", err)
	}
	if person.GivenName != "" {
		t.Errorf("GivenName = %q, want empty", person.GivenName)
	}
}

func TestParsePerson_WithNoteRef(t *testing.T) {
	// Create event with tag < 0x03E8 containing a note reference
	// Event sub-data: 18 bytes header + sub-TLV [len=8][tag=0x0000][noteID=99]
	eventData := make([]byte, 26)
	binary.LittleEndian.PutUint16(eventData[18:], 8)      // sub-TLV len
	binary.LittleEndian.PutUint16(eventData[20:], 0x0000)  // sub-TLV tag
	binary.LittleEndian.PutUint32(eventData[22:], 99)      // note ID

	var recData []byte
	recData = append(recData, make([]byte, 6)...)
	recData = append(recData, makeTLVField(0x0100, eventData)...)

	rec := RawRecord{Type: RecordTypePerson, ID: 5, Data: recData}
	ec := reunion.NewErrorCollector(0)
	person, err := ParsePerson(rec, ec)
	if err != nil {
		t.Fatalf("ParsePerson() error = %v", err)
	}

	if len(person.NoteRefs) != 1 {
		t.Fatalf("NoteRefs count = %d, want 1", len(person.NoteRefs))
	}
	if person.NoteRefs[0].NoteID != 99 {
		t.Errorf("NoteRefs[0].NoteID = %d, want 99", person.NoteRefs[0].NoteID)
	}
}
