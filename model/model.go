// Package model defines the data structures for a parsed Reunion family file.
package model

// FamilyFile is the top-level aggregate of all data parsed from a Reunion bundle.
type FamilyFile struct {
	Signature        string             `json:"signature"`
	Version          int                `json:"version"`
	Header           *Header            `json:"header,omitempty"`
	Persons          []Person           `json:"persons,omitempty"`
	Families         []Family           `json:"families,omitempty"`
	Places           []Place            `json:"places,omitempty"`
	PlaceUsages      []PlaceUsage       `json:"place_usages,omitempty"`
	EventDefinitions []EventDefinition  `json:"event_definitions,omitempty"`
	Sources          []Source           `json:"sources,omitempty"`
	Notes            []Note             `json:"notes,omitempty"`
	MediaRefs        []MediaRef         `json:"media_refs,omitempty"`
	FirstNames       []FirstNameEntry   `json:"first_names,omitempty"`
	Surnames         []SurnameEntry     `json:"surnames,omitempty"`
	SearchNames      []SearchName       `json:"search_names,omitempty"`
	Timestamps       []TimestampEntry   `json:"timestamps,omitempty"`
	Bookmarks        *Bookmark          `json:"bookmarks,omitempty"`
	ColorTags        []ColorTag         `json:"color_tags,omitempty"`
	Associations     []Association      `json:"associations,omitempty"`
	FindText         string             `json:"find_text,omitempty"`
	Description      string             `json:"description,omitempty"`
	GlobalRecords    *GlobalRecordEntry `json:"global_records,omitempty"`
	Members          []Member           `json:"members,omitempty"`
	Warnings         []string           `json:"warnings,omitempty"`
}

// Header contains metadata extracted from the familydata file header.
type Header struct {
	Magic    string `json:"magic"`
	DeviceID string `json:"device_id,omitempty"`
	Model    string `json:"model,omitempty"`
	Serial   string `json:"serial,omitempty"`
	AppPath  string `json:"app_path,omitempty"`
}
