package reunion_test

import (
	"encoding/json"
	"testing"

	reunion "github.com/kevin-cantwell/reunion-go"
	_ "github.com/kevin-cantwell/reunion-go/parser" // register v14 parser
)

const testBundle = "/Users/kevin/Downloads/Cantwell 14.familyfile14"

func TestOpen(t *testing.T) {
	ff, err := reunion.Open(testBundle, nil)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	// Verify signature
	if ff.Signature != "404065" {
		t.Errorf("Signature = %q, want %q", ff.Signature, "404065")
	}

	// Verify version
	if ff.Version != 14 {
		t.Errorf("Version = %d, want 14", ff.Version)
	}

	// Verify header
	if ff.Header == nil {
		t.Fatal("Header is nil")
	}
	if ff.Header.Magic != "3SDUAU~R" {
		t.Errorf("Header.Magic = %q, want %q", ff.Header.Magic, "3SDUAU~R")
	}

	// Verify places from cache
	if len(ff.Places) < 487 {
		t.Errorf("Places = %d, want >= 487", len(ff.Places))
	}

	// Verify first names from cache (353 in file, last record truncated = 352 parseable)
	if len(ff.FirstNames) < 352 {
		t.Errorf("FirstNames = %d, want >= 352", len(ff.FirstNames))
	}

	// Verify persons from familydata
	if len(ff.Persons) < 1094 {
		t.Errorf("Persons = %d, want >= 1094", len(ff.Persons))
	}

	// Verify most persons have names (TLV fix)
	namedCount := 0
	for _, p := range ff.Persons {
		if p.GivenName != "" || p.Surname != "" {
			namedCount++
		}
	}
	if namedCount < 1000 {
		t.Errorf("Named persons = %d, want >= 1000", namedCount)
	}
	t.Logf("Named persons: %d / %d", namedCount, len(ff.Persons))

	// Verify families from familydata
	if len(ff.Families) < 770 {
		t.Errorf("Families = %d, want >= 770", len(ff.Families))
	}

	// Verify most families have at least one partner (TLV fix)
	partnerCount := 0
	for _, f := range ff.Families {
		if f.Partner1 > 0 || f.Partner2 > 0 {
			partnerCount++
		}
	}
	if partnerCount < 700 {
		t.Errorf("Families with partners = %d, want >= 700", partnerCount)
	}
	t.Logf("Families with partners: %d / %d", partnerCount, len(ff.Families))

	// Verify event definitions
	if len(ff.EventDefinitions) < 162 {
		t.Errorf("EventDefinitions = %d, want >= 162", len(ff.EventDefinitions))
	}

	// Verify note files discovered
	noteFileCount := 0
	for _, n := range ff.Notes {
		if n.Filename != "" {
			noteFileCount++
		}
	}
	if noteFileCount < 30 {
		t.Errorf("Note files = %d, want >= 30", noteFileCount)
	}

	// Verify timestamps
	if len(ff.Timestamps) < 153 {
		t.Errorf("Timestamps = %d, want >= 153", len(ff.Timestamps))
	}

	// Verify place usages
	if len(ff.PlaceUsages) < 487 {
		t.Errorf("PlaceUsages = %d, want >= 487", len(ff.PlaceUsages))
	}

	// Verify JSON round-trip
	jsonData, err := ff.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Fatal("ToJSON() produced empty output")
	}

	// Verify JSON is valid
	var raw json.RawMessage
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("JSON output is not valid: %v", err)
	}

	// Verify compact JSON
	compactData, err := ff.ToJSONCompact()
	if err != nil {
		t.Fatalf("ToJSONCompact() error: %v", err)
	}
	if len(compactData) >= len(jsonData) {
		t.Errorf("Compact JSON (%d bytes) should be smaller than indented (%d bytes)", len(compactData), len(jsonData))
	}

	// Log some stats
	t.Logf("Persons: %d", len(ff.Persons))
	t.Logf("Families: %d", len(ff.Families))
	t.Logf("Places: %d", len(ff.Places))
	t.Logf("PlaceUsages: %d", len(ff.PlaceUsages))
	t.Logf("EventDefinitions: %d", len(ff.EventDefinitions))
	t.Logf("FirstNames: %d", len(ff.FirstNames))
	t.Logf("Note files: %d", noteFileCount)
	t.Logf("Inline notes: %d", len(ff.Notes)-noteFileCount)
	t.Logf("Sources: %d", len(ff.Sources))
	t.Logf("MediaRefs: %d", len(ff.MediaRefs))
	t.Logf("Timestamps: %d", len(ff.Timestamps))
	t.Logf("Warnings: %d", len(ff.Warnings))
	t.Logf("JSON size: %d bytes", len(jsonData))
}
