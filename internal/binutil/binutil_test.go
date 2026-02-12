package binutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestU16LE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int
		want    uint16
		wantErr bool
	}{
		{"zero", []byte{0x00, 0x00}, 0, 0, false},
		{"max", []byte{0xFF, 0xFF}, 0, 0xFFFF, false},
		{"little-endian", []byte{0x01, 0x02}, 0, 0x0201, false},
		{"with offset", []byte{0xAA, 0x34, 0x12}, 1, 0x1234, false},
		{"short data", []byte{0x01}, 0, 0, true},
		{"offset past end", []byte{0x01, 0x02}, 1, 0, true},
		{"empty", []byte{}, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := U16LE(tt.data, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("U16LE() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrShortRead) {
				t.Errorf("U16LE() error should wrap ErrShortRead, got %v", err)
			}
			if got != tt.want {
				t.Errorf("U16LE() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestU32LE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int
		want    uint32
		wantErr bool
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, 0, false},
		{"max", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF, false},
		{"little-endian", []byte{0x78, 0x56, 0x34, 0x12}, 0, 0x12345678, false},
		{"with offset", []byte{0xAA, 0x01, 0x00, 0x00, 0x00}, 1, 1, false},
		{"short data", []byte{0x01, 0x02, 0x03}, 0, 0, true},
		{"empty", []byte{}, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := U32LE(tt.data, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("U32LE() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrShortRead) {
				t.Errorf("U32LE() error should wrap ErrShortRead, got %v", err)
			}
			if got != tt.want {
				t.Errorf("U32LE() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestReadU16LE(t *testing.T) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], 0xABCD)
	got, err := ReadU16LE(bytes.NewReader(buf[:]))
	if err != nil {
		t.Fatalf("ReadU16LE() error = %v", err)
	}
	if got != 0xABCD {
		t.Errorf("ReadU16LE() = 0x%04X, want 0xABCD", got)
	}

	// Empty reader should error
	_, err = ReadU16LE(bytes.NewReader(nil))
	if err == nil {
		t.Error("ReadU16LE(empty) should error")
	}
}

func TestReadU32LE(t *testing.T) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], 0xDEADBEEF)
	got, err := ReadU32LE(bytes.NewReader(buf[:]))
	if err != nil {
		t.Fatalf("ReadU32LE() error = %v", err)
	}
	if got != 0xDEADBEEF {
		t.Errorf("ReadU32LE() = 0x%08X, want 0xDEADBEEF", got)
	}

	// Empty reader should error
	_, err = ReadU32LE(bytes.NewReader(nil))
	if err == nil {
		t.Error("ReadU32LE(empty) should error")
	}
}

func TestReadNullTermString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int
		wantStr  string
		wantSize int
	}{
		{"simple", []byte("hello\x00world"), 0, "hello", 6},
		{"with offset", []byte("XXhello\x00"), 2, "hello", 6},
		{"empty string", []byte{0x00, 0x41}, 0, "", 1},
		{"no terminator", []byte("hello"), 0, "hello", 5},
		{"mid-data", []byte("ab\x00cd\x00"), 0, "ab", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotSize := ReadNullTermString(tt.data, tt.offset)
			if gotStr != tt.wantStr {
				t.Errorf("ReadNullTermString() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotSize != tt.wantSize {
				t.Errorf("ReadNullTermString() size = %d, want %d", gotSize, tt.wantSize)
			}
		})
	}
}

func TestReadNewlineTermString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int
		wantStr  string
		wantSize int
	}{
		{"simple", []byte("hello\nworld"), 0, "hello", 6},
		{"with offset", []byte("XXhi\n"), 2, "hi", 3},
		{"empty string", []byte("\nfoo"), 0, "", 1},
		{"no terminator", []byte("hello"), 0, "hello", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotSize := ReadNewlineTermString(tt.data, tt.offset)
			if gotStr != tt.wantStr {
				t.Errorf("ReadNewlineTermString() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotSize != tt.wantSize {
				t.Errorf("ReadNewlineTermString() size = %d, want %d", gotSize, tt.wantSize)
			}
		})
	}
}

func TestReadLenPrefixedString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int
		wantStr  string
		wantSize int
		wantErr  bool
	}{
		{"simple", []byte{5, 'h', 'e', 'l', 'l', 'o'}, 0, "hello", 6, false},
		{"with offset", []byte{0xFF, 2, 'h', 'i'}, 1, "hi", 3, false},
		{"empty string", []byte{0}, 0, "", 1, false},
		{"past end", []byte{}, 0, "", 0, true},
		{"length exceeds data", []byte{10, 'h', 'i'}, 0, "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotSize, err := ReadLenPrefixedString(tt.data, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadLenPrefixedString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if gotStr != tt.wantStr {
				t.Errorf("ReadLenPrefixedString() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotSize != tt.wantSize {
				t.Errorf("ReadLenPrefixedString() size = %d, want %d", gotSize, tt.wantSize)
			}
		})
	}
}
