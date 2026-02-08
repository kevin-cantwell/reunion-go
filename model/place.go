package model

// Place represents a place from the places cache or familydata.
type Place struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
	Ref  []byte `json:"-"`
}

// PlaceUsage links a place to entities that reference it.
type PlaceUsage struct {
	PlaceID uint32             `json:"place_id"`
	Entries []PlaceUsageEntry  `json:"entries,omitempty"`
}

// PlaceUsageEntry is a single reference from an entity to a place.
type PlaceUsageEntry struct {
	RefID    uint32 `json:"ref_id"`
	TypeCode uint32 `json:"type_code"`
}
