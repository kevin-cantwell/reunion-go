package familydata

import (
	"encoding/binary"
	"testing"
)

// makeRecord builds a raw binary record suitable for ScanRecords.
// Layout:
//
//	[4 bytes padding] [2 bytes seqNum LE] [2 bytes typeCode LE]
//	[4 bytes marker 05030201] [4 bytes dataLen LE] [4 bytes recordID LE]
//	[4 bytes timestamp] [data...]
func makeRecord(seqNum uint16, typeCode RecordType, id uint32, data []byte) []byte {
	hdrSize := 20 // 4 pad + 2 seq + 2 type + 4 marker + 4 dataLen + 4 id
	tsSize := 4   // timestamp (part of data area in record layout)
	buf := make([]byte, hdrSize+tsSize+len(data))

	// padding (4 bytes of zeros)
	// seqNum at offset 4
	binary.LittleEndian.PutUint16(buf[4:], seqNum)
	// typeCode at offset 6
	binary.LittleEndian.PutUint16(buf[6:], uint16(typeCode))
	// marker at offset 8
	copy(buf[8:12], Marker)
	// dataLen at offset 12 — includes timestamp + data
	binary.LittleEndian.PutUint32(buf[12:], uint32(tsSize+len(data)))
	// recordID at offset 16
	binary.LittleEndian.PutUint32(buf[16:], id)
	// timestamp at offset 20 (zeros)
	// data at offset 24
	copy(buf[24:], data)
	return buf
}

func TestScanRecords_SingleRecord(t *testing.T) {
	payload := []byte("test data payload")
	raw := makeRecord(1, RecordTypePerson, 100, payload)
	records := ScanRecords(raw)

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.Type != RecordTypePerson {
		t.Errorf("Type = 0x%04X, want 0x%04X", rec.Type, RecordTypePerson)
	}
	if rec.SeqNum != 1 {
		t.Errorf("SeqNum = %d, want 1", rec.SeqNum)
	}
	if rec.ID != 100 {
		t.Errorf("ID = %d, want 100", rec.ID)
	}
}

func TestScanRecords_MultipleRecords(t *testing.T) {
	rec1 := makeRecord(1, RecordTypePerson, 10, []byte("person"))
	rec2 := makeRecord(2, RecordTypeFamily, 20, []byte("family"))
	rec3 := makeRecord(3, RecordTypeSchema, 30, []byte("schema"))

	var data []byte
	data = append(data, rec1...)
	data = append(data, rec2...)
	data = append(data, rec3...)

	records := ScanRecords(data)
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	expected := []struct {
		typ RecordType
		id  uint32
		seq uint16
	}{
		{RecordTypePerson, 10, 1},
		{RecordTypeFamily, 20, 2},
		{RecordTypeSchema, 30, 3},
	}

	for i, e := range expected {
		if records[i].Type != e.typ {
			t.Errorf("record[%d].Type = 0x%04X, want 0x%04X", i, records[i].Type, e.typ)
		}
		if records[i].ID != e.id {
			t.Errorf("record[%d].ID = %d, want %d", i, records[i].ID, e.id)
		}
		if records[i].SeqNum != e.seq {
			t.Errorf("record[%d].SeqNum = %d, want %d", i, records[i].SeqNum, e.seq)
		}
	}
}

func TestScanRecords_OverflowExtension(t *testing.T) {
	// Create two records where the first has overflow data between its
	// declared DataLen boundary and the next record's marker.
	rec1 := makeRecord(1, RecordTypePerson, 10, []byte("short"))
	overflow := []byte{0xAA, 0xBB, 0xCC, 0xDD} // extra bytes
	rec2 := makeRecord(2, RecordTypeFamily, 20, []byte("next"))

	var data []byte
	data = append(data, rec1...)
	data = append(data, overflow...)
	data = append(data, rec2...)

	records := ScanRecords(data)
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// The first record's Data should be extended to include the overflow
	// bytes up to the next record's marker position.
	rec1DataStart := records[0].Offset + 20
	rec2MarkerPos := records[1].Offset + 8
	expectedDataLen := rec2MarkerPos - rec1DataStart
	if len(records[0].Data) != expectedDataLen {
		t.Errorf("record[0].Data len = %d, want %d (with overflow)", len(records[0].Data), expectedDataLen)
	}
}

func TestScanRecords_NoMarker(t *testing.T) {
	data := []byte("no marker here at all, just regular data bytes")
	records := ScanRecords(data)
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}
