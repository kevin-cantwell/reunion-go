// Package index provides fast lookup structures for a parsed FamilyFile.
package index

import (
	"strings"

	"github.com/kevin-cantwell/reunion-explore/model"
)

// TreeEntry pairs a person with their generation depth.
type TreeEntry struct {
	Generation int           `json:"generation"`
	Person     *model.Person `json:"person"`
}

// Index provides fast lookups into a parsed FamilyFile.
type Index struct {
	Persons         map[uint32]*model.Person
	Families        map[uint32]*model.Family
	Places          map[uint32]*model.Place
	Schemas         map[uint32]*model.EventDefinition
	Sources         map[uint32]*model.Source
	Notes           map[uint32]*model.Note
	ChildFamilies   map[uint32][]uint32 // personID -> familyIDs where they're a child
	PartnerFamilies map[uint32][]uint32   // personID -> familyIDs where they're a partner
	SurnameIndex    map[string][]uint32   // lowercase surname -> personIDs
	PlacePersons    map[uint32][]uint32   // placeID -> personIDs with events at that place
	SchemaPersons   map[uint32][]uint32   // schemaID -> personIDs with that event type
}

// BuildIndex creates lookup indexes from a parsed FamilyFile.
func BuildIndex(ff *model.FamilyFile) *Index {
	idx := &Index{
		Persons:         make(map[uint32]*model.Person, len(ff.Persons)),
		Families:        make(map[uint32]*model.Family, len(ff.Families)),
		Places:          make(map[uint32]*model.Place, len(ff.Places)),
		Schemas:         make(map[uint32]*model.EventDefinition, len(ff.EventDefinitions)),
		Sources:         make(map[uint32]*model.Source, len(ff.Sources)),
		Notes:           make(map[uint32]*model.Note, len(ff.Notes)),
		ChildFamilies:   make(map[uint32][]uint32),
		PartnerFamilies: make(map[uint32][]uint32),
		SurnameIndex:    make(map[string][]uint32),
		PlacePersons:    make(map[uint32][]uint32),
		SchemaPersons:   make(map[uint32][]uint32),
	}

	for i := range ff.Persons {
		p := &ff.Persons[i]
		idx.Persons[p.ID] = p

		if p.Surname != "" {
			key := strings.ToLower(p.Surname)
			idx.SurnameIndex[key] = append(idx.SurnameIndex[key], p.ID)
		}

		for _, evt := range p.Events {
			if evt.SchemaID > 0 {
				sid := uint32(evt.SchemaID)
				idx.SchemaPersons[sid] = appendUnique(idx.SchemaPersons[sid], p.ID)
			}
			for _, placeRef := range evt.PlaceRefs {
				pid := uint32(placeRef)
				idx.PlacePersons[pid] = appendUnique(idx.PlacePersons[pid], p.ID)
			}
		}
	}

	for i := range ff.Families {
		f := &ff.Families[i]
		idx.Families[f.ID] = f

		if f.Partner1 > 0 {
			idx.PartnerFamilies[f.Partner1] = append(idx.PartnerFamilies[f.Partner1], f.ID)
		}
		if f.Partner2 > 0 {
			idx.PartnerFamilies[f.Partner2] = append(idx.PartnerFamilies[f.Partner2], f.ID)
		}
		for _, childID := range f.Children {
			idx.ChildFamilies[childID] = append(idx.ChildFamilies[childID], f.ID)
		}
	}

	for i := range ff.Places {
		idx.Places[ff.Places[i].ID] = &ff.Places[i]
	}

	for i := range ff.EventDefinitions {
		idx.Schemas[ff.EventDefinitions[i].ID] = &ff.EventDefinitions[i]
	}

	for i := range ff.Sources {
		idx.Sources[ff.Sources[i].ID] = &ff.Sources[i]
	}

	for i := range ff.Notes {
		idx.Notes[ff.Notes[i].ID] = &ff.Notes[i]
	}

	return idx
}

func appendUnique(slice []uint32, val uint32) []uint32 {
	for _, v := range slice {
		if v == val {
			return slice
		}
	}
	return append(slice, val)
}

// FormatName returns "Given Surname" for a person.
func FormatName(p *model.Person) string {
	if p.GivenName == "" && p.Surname == "" {
		return "(unknown)"
	}
	if p.GivenName == "" {
		return p.Surname
	}
	if p.Surname == "" {
		return p.GivenName
	}
	return p.GivenName + " " + p.Surname
}

// PersonName returns a formatted name for a person ID.
func (idx *Index) PersonName(id uint32) string {
	p, ok := idx.Persons[id]
	if !ok {
		return "?"
	}
	return FormatName(p)
}

// PlaceName returns the name for a place ID.
func (idx *Index) PlaceName(id int) string {
	p, ok := idx.Places[uint32(id)]
	if !ok {
		return ""
	}
	return p.Name
}

// SchemaName returns the display name for a schema/event definition ID.
func (idx *Index) SchemaName(id uint16) string {
	s, ok := idx.Schemas[uint32(id)]
	if !ok {
		return ""
	}
	return s.DisplayName
}

// Parents returns the partner IDs from families where personID is a child.
func (idx *Index) Parents(personID uint32) (parents []uint32) {
	for _, famID := range idx.ChildFamilies[personID] {
		f, ok := idx.Families[famID]
		if !ok {
			continue
		}
		if f.Partner1 > 0 {
			parents = append(parents, f.Partner1)
		}
		if f.Partner2 > 0 {
			parents = append(parents, f.Partner2)
		}
	}
	return
}

// Siblings returns sibling IDs (other children in the same family).
func (idx *Index) Siblings(personID uint32) []uint32 {
	seen := map[uint32]bool{personID: true}
	var siblings []uint32
	for _, famID := range idx.ChildFamilies[personID] {
		f, ok := idx.Families[famID]
		if !ok {
			continue
		}
		for _, childID := range f.Children {
			if !seen[childID] {
				seen[childID] = true
				siblings = append(siblings, childID)
			}
		}
	}
	return siblings
}

// Spouses returns partner IDs from families where personID is a partner.
func (idx *Index) Spouses(personID uint32) []uint32 {
	seen := map[uint32]bool{personID: true}
	var spouses []uint32
	for _, famID := range idx.PartnerFamilies[personID] {
		f, ok := idx.Families[famID]
		if !ok {
			continue
		}
		if f.Partner1 > 0 && !seen[f.Partner1] {
			seen[f.Partner1] = true
			spouses = append(spouses, f.Partner1)
		}
		if f.Partner2 > 0 && !seen[f.Partner2] {
			seen[f.Partner2] = true
			spouses = append(spouses, f.Partner2)
		}
	}
	return spouses
}

// ChildrenOf returns child IDs from families where personID is a partner.
func (idx *Index) ChildrenOf(personID uint32) []uint32 {
	seen := make(map[uint32]bool)
	var children []uint32
	for _, famID := range idx.PartnerFamilies[personID] {
		f, ok := idx.Families[famID]
		if !ok {
			continue
		}
		for _, childID := range f.Children {
			if !seen[childID] {
				seen[childID] = true
				children = append(children, childID)
			}
		}
	}
	return children
}

// PersonsBySurname returns person IDs matching the given surname (case-insensitive).
func (idx *Index) PersonsBySurname(surname string) []uint32 {
	return idx.SurnameIndex[strings.ToLower(surname)]
}

// PersonsByPlace returns person IDs who have events at the given place.
func (idx *Index) PersonsByPlace(placeID uint32) []uint32 {
	return idx.PlacePersons[placeID]
}

// PersonsBySchema returns person IDs who have the given event type.
func (idx *Index) PersonsBySchema(schemaID uint32) []uint32 {
	return idx.SchemaPersons[schemaID]
}

// FamiliesForPerson returns all family IDs where the person is a partner or child.
func (idx *Index) FamiliesForPerson(personID uint32) []uint32 {
	seen := make(map[uint32]bool)
	var fams []uint32
	for _, fid := range idx.PartnerFamilies[personID] {
		if !seen[fid] {
			seen[fid] = true
			fams = append(fams, fid)
		}
	}
	for _, fid := range idx.ChildFamilies[personID] {
		if !seen[fid] {
			seen[fid] = true
			fams = append(fams, fid)
		}
	}
	return fams
}

// Ancestors collects ancestors up to maxGen generations deep.
func (idx *Index) Ancestors(id uint32, maxGen int) []TreeEntry {
	var entries []TreeEntry
	visited := make(map[uint32]bool)
	idx.collectAncestors(id, 0, maxGen, visited, &entries)
	return entries
}

func (idx *Index) collectAncestors(id uint32, gen, maxGen int, visited map[uint32]bool, out *[]TreeEntry) {
	if gen > maxGen || visited[id] {
		return
	}
	visited[id] = true

	for _, pid := range idx.Parents(id) {
		if visited[pid] {
			continue
		}
		p, ok := idx.Persons[pid]
		if !ok {
			continue
		}
		*out = append(*out, TreeEntry{Generation: gen + 1, Person: p})
		idx.collectAncestors(pid, gen+1, maxGen, visited, out)
	}
}

// Descendants collects descendants up to maxGen generations deep.
func (idx *Index) Descendants(id uint32, maxGen int) []TreeEntry {
	var entries []TreeEntry
	visited := make(map[uint32]bool)
	idx.collectDescendants(id, 0, maxGen, visited, &entries)
	return entries
}

func (idx *Index) collectDescendants(id uint32, gen, maxGen int, visited map[uint32]bool, out *[]TreeEntry) {
	if gen > maxGen || visited[id] {
		return
	}
	visited[id] = true

	for _, cid := range idx.ChildrenOf(id) {
		if visited[cid] {
			continue
		}
		c, ok := idx.Persons[cid]
		if !ok {
			continue
		}
		*out = append(*out, TreeEntry{Generation: gen + 1, Person: c})
		idx.collectDescendants(cid, gen+1, maxGen, visited, out)
	}
}

// Treetops finds terminal ancestors (persons with no parents in the dataset).
func (idx *Index) Treetops(id uint32) []*model.Person {
	visited := make(map[uint32]bool)
	var treetopIDs []uint32
	idx.findTreetops(id, visited, &treetopIDs)

	var persons []*model.Person
	for _, tid := range treetopIDs {
		if p, ok := idx.Persons[tid]; ok {
			persons = append(persons, p)
		}
	}
	return persons
}

func (idx *Index) findTreetops(id uint32, visited map[uint32]bool, treetops *[]uint32) {
	if visited[id] {
		return
	}
	visited[id] = true

	parents := idx.Parents(id)
	if len(parents) == 0 {
		*treetops = append(*treetops, id)
		return
	}
	for _, pid := range parents {
		idx.findTreetops(pid, visited, treetops)
	}
}

// Search finds persons whose name contains the query (case-insensitive).
func (idx *Index) Search(query string) []*model.Person {
	q := strings.ToLower(query)
	var matches []*model.Person
	for _, p := range idx.Persons {
		name := strings.ToLower(FormatName(p))
		if strings.Contains(name, q) {
			matches = append(matches, p)
		}
	}
	return matches
}
