package familydata

import (
	"encoding/binary"
	"testing"
)

// putU16LE writes a little-endian uint16 into buf at the given offset.
func putU16LE(buf []byte, offset int, v uint16) {
	binary.LittleEndian.PutUint16(buf[offset:], v)
}

// putU32LE writes a little-endian uint32 into buf at the given offset.
func putU32LE(buf []byte, offset int, v uint32) {
	binary.LittleEndian.PutUint32(buf[offset:], v)
}

// makeTLVField builds a single TLV field: [totalLen u16LE][tag u16LE][data...].
func makeTLVField(tag uint16, data []byte) []byte {
	totalLen := uint16(len(data) + 4)
	buf := make([]byte, totalLen)
	putU16LE(buf, 0, totalLen)
	putU16LE(buf, 2, tag)
	copy(buf[4:], data)
	return buf
}

// makePreamble creates the 6-byte preamble: 4-byte timestamp + 2-byte size.
func makePreamble() []byte {
	return make([]byte, 6)
}

func TestParseTLVFields(t *testing.T) {
	// Build record data: 6-byte preamble + two TLV fields
	field1Data := []byte("Alice")
	field2Data := []byte{0x01}

	var data []byte
	data = append(data, makePreamble()...)
	data = append(data, makeTLVField(0x001E, field1Data)...) // given name
	data = append(data, makeTLVField(0x001B, field2Data)...) // sex

	fields := ParseTLVFields(data)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	if fields[0].Tag != 0x001E {
		t.Errorf("field[0].Tag = 0x%04X, want 0x001E", fields[0].Tag)
	}
	if string(fields[0].Data) != "Alice" {
		t.Errorf("field[0].Data = %q, want %q", fields[0].Data, "Alice")
	}

	if fields[1].Tag != 0x001B {
		t.Errorf("field[1].Tag = 0x%04X, want 0x001B", fields[1].Tag)
	}
	if len(fields[1].Data) != 1 || fields[1].Data[0] != 0x01 {
		t.Errorf("field[1].Data = %v, want [0x01]", fields[1].Data)
	}
}

func TestParseTLVFields_Empty(t *testing.T) {
	// Too short for preamble
	if fields := ParseTLVFields(nil); fields != nil {
		t.Errorf("expected nil for nil data, got %v", fields)
	}
	if fields := ParseTLVFields([]byte{0, 0, 0}); fields != nil {
		t.Errorf("expected nil for short data, got %v", fields)
	}
}

func TestParseTLVFields_TotalLenLessThan4Stops(t *testing.T) {
	// A field with totalLen < 4 should stop parsing (used as sentinel)
	var data []byte
	data = append(data, makePreamble()...)
	data = append(data, makeTLVField(0x0001, []byte("ok"))...)
	// Append a zero-length sentinel (totalLen=0)
	data = append(data, 0, 0, 0, 0)
	data = append(data, makeTLVField(0x0002, []byte("hidden"))...)

	fields := ParseTLVFields(data)
	if len(fields) != 1 {
		t.Fatalf("expected 1 field (sentinel should stop), got %d", len(fields))
	}
	if fields[0].Tag != 0x0001 {
		t.Errorf("field[0].Tag = 0x%04X, want 0x0001", fields[0].Tag)
	}
}

func TestExtractPlaceRefs(t *testing.T) {
	data := []byte("born in [[pt:42]] and died in [[pt:99]]")
	refs := ExtractPlaceRefs(data)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0] != 42 {
		t.Errorf("refs[0] = %d, want 42", refs[0])
	}
	if refs[1] != 99 {
		t.Errorf("refs[1] = %d, want 99", refs[1])
	}

	// No matches
	none := ExtractPlaceRefs([]byte("no refs here"))
	if len(none) != 0 {
		t.Errorf("expected 0 refs, got %d", len(none))
	}
}

func TestExtractString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"printable ascii", []byte{0x00, 0x01, 'H', 'e', 'l', 'l', 'o', 0x00}, "Hello"},
		{"all printable", []byte("Test 123"), "Test 123"},
		{"leading non-printable", []byte{0x00, 0x01, 0x02, 'A', 'B'}, "AB"},
		{"utf8 multi-byte", append([]byte{0x00}, []byte("café")...), "café"},
		{"empty", []byte{0x00, 0x01, 0x02}, ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractString(tt.data)
			if got != tt.want {
				t.Errorf("ExtractString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractDate(t *testing.T) {
	// Helper to build a 26-byte event sub-data with date encoding
	buildDateData := func(precFlags, dayByte byte, hdrVal uint16, year, quarter int) []byte {
		buf := make([]byte, 26)
		// hdrVal at offset 2 (u16 LE) — encodes month within quarter
		putU16LE(buf, 2, hdrVal)
		// Sub-TLV at offset 18: length 8 means date sub-TLV
		putU16LE(buf, 18, 8)
		// Precision flags at offset 22
		buf[22] = precFlags
		// Day byte at offset 23
		buf[23] = dayByte
		// Year+quarter encoding at offset 24 (u16 LE)
		totalQ := uint16((year+8000)*4 + quarter)
		putU16LE(buf, 24, totalQ)
		return buf
	}

	tests := []struct {
		name      string
		data      []byte
		want      string
	}{
		{
			name: "normal date: 15 Mar 1990",
			// quarter=0 (Q1=Jan-Mar), hdrVal%3=0 → month offset 0 → month=1 (Jan)
			// Wait: quarter 0 = Q1(Jan-Mar), monthInQuarter depends on hdrVal%3
			// hdrVal%3=0 → offset 0 → month 1; hdrVal%3=1 → offset 2 → month 3; hdrVal%3=2 → offset 1 → month 2
			// For March: quarter=0, monthOffset=2 → hdrVal%3=1
			data: buildDateData(0x00, 15, 1, 1990, 0),
			want: "15 Mar 1990",
		},
		{
			name: "month only: Jun 2000",
			// June = quarter 1 (Apr-Jun), month offset 2 (Jun is 3rd in Q2) → hdrVal%3=1
			data: buildDateData(0x00, 0, 1, 2000, 1),
			want: "Jun 2000",
		},
		{
			name: "about year only: about 1850",
			data: buildDateData(0xA0, 0, 0, 1850, 0),
			want: "about 1850",
		},
		{
			name: "after date: after 1 Jan 1900",
			// Jan = quarter 0, monthOffset 0, hdrVal%3=0
			data: buildDateData(0x40, 1, 0, 1900, 0),
			want: "after 1 Jan 1900",
		},
		{
			name: "before date: before 25 Dec 1800",
			// Dec = quarter 3, monthOffset 2 (Dec is 3rd in Q4) → hdrVal%3=1
			data: buildDateData(0x00, 25|0xC0, 1, 1800, 3),
			want: "before 25 Dec 1800",
		},
		{
			name: "too short",
			data: make([]byte, 10),
			want: "",
		},
		{
			name: "sub-TLV length not 8",
			data: func() []byte {
				buf := make([]byte, 26)
				putU16LE(buf, 18, 6) // not 8
				return buf
			}(),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractDate(tt.data)
			if got != tt.want {
				t.Errorf("ExtractDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractNoteRef(t *testing.T) {
	// Build event sub-data with a note reference sub-TLV at offset 18
	// Sub-TLV: [totalLen=8 u16LE][tag=0x0000 u16LE][noteID u32LE]
	buf := make([]byte, 26)
	putU16LE(buf, 18, 8)      // totalLen = 8
	putU16LE(buf, 20, 0x0000) // tag = 0x0000
	putU32LE(buf, 22, 42)     // note ID = 42

	got := ExtractNoteRef(buf)
	if got != 42 {
		t.Errorf("ExtractNoteRef() = %d, want 42", got)
	}

	// No note ref: wrong tag
	buf2 := make([]byte, 26)
	putU16LE(buf2, 18, 8)
	putU16LE(buf2, 20, 0x0001) // not 0x0000
	putU32LE(buf2, 22, 42)
	got2 := ExtractNoteRef(buf2)
	if got2 != 0 {
		t.Errorf("ExtractNoteRef(wrong tag) = %d, want 0", got2)
	}

	// Too short
	got3 := ExtractNoteRef(make([]byte, 10))
	if got3 != 0 {
		t.Errorf("ExtractNoteRef(short) = %d, want 0", got3)
	}
}
