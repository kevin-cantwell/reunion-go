package familydata

import (
	"testing"
)

func TestParseHeader(t *testing.T) {
	// Build a header buffer: magic at offset 0, then newline-terminated
	// strings starting at a printable area after offset 80.
	buf := make([]byte, 256)
	copy(buf[0:8], "3SDUAU~R")

	// Place newline-terminated strings starting at offset 80
	pos := 80
	devID := "DEVICE123"
	copy(buf[pos:], devID)
	pos += len(devID)
	buf[pos] = '\n'
	pos++

	modelName := "MacBookPro"
	copy(buf[pos:], modelName)
	pos += len(modelName)
	buf[pos] = '\n'
	pos++

	serial := "SN-12345"
	copy(buf[pos:], serial)
	pos += len(serial)
	buf[pos] = '\n'
	pos++

	appPath := "/Applications/Reunion.app"
	copy(buf[pos:], appPath)
	pos += len(appPath)
	buf[pos] = 0 // null-terminated

	h, err := ParseHeader(buf)
	if err != nil {
		t.Fatalf("ParseHeader() error = %v", err)
	}

	if h.Magic != "3SDUAU~R" {
		t.Errorf("Magic = %q, want %q", h.Magic, "3SDUAU~R")
	}
	if h.DeviceID != devID {
		t.Errorf("DeviceID = %q, want %q", h.DeviceID, devID)
	}
	if h.Model != modelName {
		t.Errorf("Model = %q, want %q", h.Model, modelName)
	}
	if h.Serial != serial {
		t.Errorf("Serial = %q, want %q", h.Serial, serial)
	}
	if h.AppPath != appPath {
		t.Errorf("AppPath = %q, want %q", h.AppPath, appPath)
	}
}

func TestParseHeader_TooShort(t *testing.T) {
	_, err := ParseHeader([]byte("short"))
	if err == nil {
		t.Error("ParseHeader() should error on data shorter than 8 bytes")
	}
}

func TestParseHeader_MagicOnly(t *testing.T) {
	buf := make([]byte, 8)
	copy(buf, "3SDUAU~R")
	h, err := ParseHeader(buf)
	if err != nil {
		t.Fatalf("ParseHeader() error = %v", err)
	}
	if h.Magic != "3SDUAU~R" {
		t.Errorf("Magic = %q, want %q", h.Magic, "3SDUAU~R")
	}
	// No strings area, should still return the header
	if h.DeviceID != "" {
		t.Errorf("DeviceID = %q, want empty", h.DeviceID)
	}
}
