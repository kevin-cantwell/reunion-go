package main

import "github.com/kevin-cantwell/reunion-go/model"

// Index provides fast lookups into a parsed FamilyFile.
type Index struct {
	Persons         map[uint32]*model.Person
	Families        map[uint32]*model.Family
	Places          map[uint32]*model.Place
	Schemas         map[uint32]*model.EventDefinition
	ChildFamilies   map[uint32][]uint32 // personID -> familyIDs where they're a child
	PartnerFamilies map[uint32][]uint32 // personID -> familyIDs where they're a partner
}

// BuildIndex creates lookup indexes from a parsed FamilyFile.
func BuildIndex(ff *model.FamilyFile) *Index {
	idx := &Index{
		Persons:         make(map[uint32]*model.Person, len(ff.Persons)),
		Families:        make(map[uint32]*model.Family, len(ff.Families)),
		Places:          make(map[uint32]*model.Place, len(ff.Places)),
		Schemas:         make(map[uint32]*model.EventDefinition, len(ff.EventDefinitions)),
		ChildFamilies:   make(map[uint32][]uint32),
		PartnerFamilies: make(map[uint32][]uint32),
	}

	for i := range ff.Persons {
		idx.Persons[ff.Persons[i].ID] = &ff.Persons[i]
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

	return idx
}

// PersonName returns a formatted name for a person ID.
func (idx *Index) PersonName(id uint32) string {
	p, ok := idx.Persons[id]
	if !ok {
		return "?"
	}
	return FormatName(p)
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
