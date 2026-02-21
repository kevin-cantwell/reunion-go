package familydata

import (
	"encoding/binary"
	"testing"

	reunion "github.com/kevin-cantwell/reunion-explore"
)

func TestParseFamily(t *testing.T) {
	var recData []byte
	recData = append(recData, make([]byte, 6)...) // preamble

	// Partner1 (tag 0x0050) — 4-byte ID
	p1Data := make([]byte, 4)
	binary.LittleEndian.PutUint32(p1Data, 101)
	recData = append(recData, makeTLVField(TagPartner1, p1Data)...)

	// Partner2 (tag 0x0051) — 4-byte ID
	p2Data := make([]byte, 4)
	binary.LittleEndian.PutUint32(p2Data, 102)
	recData = append(recData, makeTLVField(TagPartner2, p2Data)...)

	// Child (tag 0x00FA) — encoded as u32LE >> 8
	childID := uint32(55)
	childData := make([]byte, 4)
	binary.LittleEndian.PutUint32(childData, childID<<8)
	recData = append(recData, makeTLVField(0x00FA, childData)...)

	// Second child (tag 0x00FB)
	childID2 := uint32(56)
	childData2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(childData2, childID2<<8)
	recData = append(recData, makeTLVField(0x00FB, childData2)...)

	// Family event (tag 0x0100, >= 0x100)
	eventData := make([]byte, 20)
	recData = append(recData, makeTLVField(0x0100, eventData)...)

	rec := RawRecord{
		Type:   RecordTypeFamily,
		ID:     200,
		SeqNum: 3,
		Data:   recData,
	}

	ec := reunion.NewErrorCollector(0)
	family, err := ParseFamily(rec, ec)
	if err != nil {
		t.Fatalf("ParseFamily() error = %v", err)
	}

	if family.ID != 200 {
		t.Errorf("ID = %d, want 200", family.ID)
	}
	if family.SeqNum != 3 {
		t.Errorf("SeqNum = %d, want 3", family.SeqNum)
	}
	if family.Partner1 != 101 {
		t.Errorf("Partner1 = %d, want 101", family.Partner1)
	}
	if family.Partner2 != 102 {
		t.Errorf("Partner2 = %d, want 102", family.Partner2)
	}
	if len(family.Children) != 2 {
		t.Fatalf("Children count = %d, want 2", len(family.Children))
	}
	if family.Children[0] != 55 {
		t.Errorf("Children[0] = %d, want 55", family.Children[0])
	}
	if family.Children[1] != 56 {
		t.Errorf("Children[1] = %d, want 56", family.Children[1])
	}
	if len(family.Events) != 1 {
		t.Errorf("Events count = %d, want 1", len(family.Events))
	}
}

func TestParseFamily_Partner2ByteID(t *testing.T) {
	var recData []byte
	recData = append(recData, make([]byte, 6)...)

	// Partner with only 2-byte ID
	p1Data := make([]byte, 2)
	binary.LittleEndian.PutUint16(p1Data, 77)
	recData = append(recData, makeTLVField(TagPartner1, p1Data)...)

	rec := RawRecord{Type: RecordTypeFamily, ID: 1, Data: recData}
	ec := reunion.NewErrorCollector(0)
	family, err := ParseFamily(rec, ec)
	if err != nil {
		t.Fatalf("ParseFamily() error = %v", err)
	}
	if family.Partner1 != 77 {
		t.Errorf("Partner1 = %d, want 77", family.Partner1)
	}
}

func TestParseFamily_EmptyData(t *testing.T) {
	rec := RawRecord{Type: RecordTypeFamily, ID: 1, Data: []byte{0, 0}}
	ec := reunion.NewErrorCollector(0)
	family, err := ParseFamily(rec, ec)
	if err != nil {
		t.Fatalf("ParseFamily() error = %v", err)
	}
	if family.Partner1 != 0 || family.Partner2 != 0 {
		t.Error("expected zero partners for empty data")
	}
	if len(family.Children) != 0 {
		t.Error("expected no children for empty data")
	}
}

func TestIsChildTag(t *testing.T) {
	for tag := uint16(0x00FA); tag <= 0x00FF; tag++ {
		if !isChildTag(tag) {
			t.Errorf("isChildTag(0x%04X) = false, want true", tag)
		}
	}
	if isChildTag(0x00F9) {
		t.Error("isChildTag(0x00F9) should be false")
	}
	if isChildTag(0x0100) {
		t.Error("isChildTag(0x0100) should be false")
	}
}
