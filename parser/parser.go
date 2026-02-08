// Package parser provides the version-14 parser implementation and orchestrator.
package parser

import (
	"fmt"

	reunion "github.com/kevin-cantwell/reunion-go"
	"github.com/kevin-cantwell/reunion-go/bundle"
	"github.com/kevin-cantwell/reunion-go/model"
	"github.com/kevin-cantwell/reunion-go/parser/cache"
	"github.com/kevin-cantwell/reunion-go/parser/familydata"
	"github.com/kevin-cantwell/reunion-go/parser/member"
	"github.com/kevin-cantwell/reunion-go/parser/notes"
)

func init() {
	reunion.RegisterVersion(&V14Parser{})
}

// V14Parser implements VersionParser for Reunion 14.
type V14Parser struct{}

func (p *V14Parser) Version() reunion.Version { return reunion.Version14 }

func (p *V14Parser) CanParse(bundlePath string) (bool, error) {
	b, err := bundle.OpenBundle(bundlePath)
	if err != nil {
		return false, err
	}
	return b.FamilyData != "" && b.Signature != "", nil
}

func (p *V14Parser) Parse(bundlePath string, opts reunion.ParseOptions) (*model.FamilyFile, error) {
	b, err := bundle.OpenBundle(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("opening bundle: %w", err)
	}

	ec := reunion.NewErrorCollector(opts.MaxErrors)
	ff := &model.FamilyFile{
		Version: 14,
	}

	// Parse signature
	if b.Signature != "" {
		sig, err := ParseSignature(b.Signature)
		if err != nil {
			ec.Add("signature", -1, "failed to parse signature", err)
		} else {
			ff.Signature = sig
		}
	}

	// Parse familydata
	if b.FamilyData != "" {
		result, err := familydata.Parse(b.FamilyData, ec)
		if err != nil {
			return nil, fmt.Errorf("parsing familydata: %w", err)
		}
		ff.Header = result.Header
		ff.Persons = result.Persons
		ff.Families = result.Families
		ff.EventDefinitions = result.EventDefinitions
		ff.Sources = result.Sources
		ff.MediaRefs = result.MediaRefs
		// Inline notes from familydata
		ff.Notes = result.Notes
	}

	// Parse cache files
	if path, ok := b.Caches["places.cache"]; ok {
		places, err := cache.ParsePlaces(path)
		if err != nil {
			ec.Add("places.cache", -1, "failed to parse", err)
		} else {
			ff.Places = places
		}
	}

	if path, ok := b.Caches["placeUsage.cache"]; ok {
		usages, err := cache.ParsePlaceUsage(path)
		if err != nil {
			ec.Add("placeUsage.cache", -1, "failed to parse", err)
		} else {
			ff.PlaceUsages = usages
		}
	}

	if path, ok := b.Caches["fmnames.cache"]; ok {
		names, err := cache.ParseFmNames(path)
		if err != nil {
			ec.Add("fmnames.cache", -1, "failed to parse", err)
		} else {
			ff.FirstNames = names
		}
	}

	if path, ok := b.Caches["surnames.cache"]; ok {
		entries, err := cache.ParseSurnames(path)
		if err != nil {
			ec.Add("surnames.cache", -1, "failed to parse", err)
		} else {
			ff.Surnames = entries
		}
	}

	if path, ok := b.Caches["shNames.cache"]; ok {
		names, err := cache.ParseShNames(path)
		if err != nil {
			ec.Add("shNames.cache", -1, "failed to parse", err)
		} else {
			ff.SearchNames = names
		}
	}

	if path, ok := b.Caches["timestamps.cache"]; ok {
		entries, err := cache.ParseTimestamps(path)
		if err != nil {
			ec.Add("timestamps.cache", -1, "failed to parse", err)
		} else {
			ff.Timestamps = entries
		}
	}

	if path, ok := b.Caches["globalRecords.cache"]; ok {
		entry, err := cache.ParseGlobalRecords(path)
		if err != nil {
			ec.Add("globalRecords.cache", -1, "failed to parse", err)
		} else {
			ff.GlobalRecords = entry
		}
	}

	if path, ok := b.Caches["bookmarks.cache"]; ok {
		bk, err := cache.ParseBookmarks(path)
		if err != nil {
			ec.Add("bookmarks.cache", -1, "failed to parse", err)
		} else {
			ff.Bookmarks = bk
		}
	}

	if path, ok := b.Caches["colortags.cache"]; ok {
		tags, err := cache.ParseColorTags(path)
		if err != nil {
			ec.Add("colortags.cache", -1, "failed to parse", err)
		} else {
			ff.ColorTags = tags
		}
	}

	if path, ok := b.Caches["associations.cache"]; ok {
		assocs, err := cache.ParseAssociations(path)
		if err != nil {
			ec.Add("associations.cache", -1, "failed to parse", err)
		} else {
			ff.Associations = assocs
		}
	}

	if path, ok := b.Caches["find.cache"]; ok {
		text, err := cache.ParseFind(path)
		if err != nil {
			ec.Add("find.cache", -1, "failed to parse", err)
		} else {
			ff.FindText = text
		}
	}

	if path, ok := b.Caches["descriptions.cache"]; ok {
		desc, err := cache.ParseDescriptions(path)
		if err != nil {
			ec.Add("descriptions.cache", -1, "failed to parse", err)
		} else {
			ff.Description = desc
		}
	}

	// Parse note files from all members
	if len(b.NoteFiles) > 0 {
		fileNotes, err := notes.ParseAllNotes(b.NoteFiles)
		if err != nil {
			ec.Add("notes", -1, "failed to parse note files", err)
		} else {
			ff.Notes = append(ff.Notes, fileNotes...)
		}
	}

	// Parse members
	if len(b.Members) > 0 {
		members, err := member.ParseMembers(b.Members)
		if err != nil {
			ec.Add("members", -1, "failed to parse members", err)
		} else {
			ff.Members = members
		}
	}

	// Collect warnings
	for _, pe := range ec.Errors() {
		ff.Warnings = append(ff.Warnings, pe.Error())
	}

	return ff, nil
}
