package model

import "encoding/json"

// ToJSON serializes the FamilyFile to indented JSON.
func (f *FamilyFile) ToJSON() ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// ToJSONCompact serializes the FamilyFile to compact JSON.
func (f *FamilyFile) ToJSONCompact() ([]byte, error) {
	return json.Marshal(f)
}
