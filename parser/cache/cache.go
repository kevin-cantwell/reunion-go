// Package cache implements parsers for Reunion's .cache files.
package cache

import (
	"fmt"

	"github.com/kevin-cantwell/reunion-go/internal/binutil"
)

// CacheHeader represents the common header found in most cache files.
type CacheHeader struct {
	FileSize uint32
	Magic    string
	Count    uint32
	Extra    []byte // any extra header bytes after count
}

// ParseHeader reads a standard cache header: size(4) + magic(4) + count(4).
func ParseHeader(data []byte, magicLen int) (*CacheHeader, error) {
	if len(data) < 4+magicLen+4 {
		return nil, fmt.Errorf("cache file too short: %d bytes", len(data))
	}
	size, _ := binutil.U32LE(data, 0)
	magic := string(data[4 : 4+magicLen])
	countOff := 4 + magicLen
	count, _ := binutil.U32LE(data, countOff)

	h := &CacheHeader{
		FileSize: size,
		Magic:    magic,
		Count:    count,
	}

	extraStart := countOff + 4
	if extraStart < len(data) {
		// Grab remaining header bytes up to 4 more
		end := extraStart + 4
		if end > len(data) {
			end = len(data)
		}
		h.Extra = data[extraStart:end]
	}

	return h, nil
}

// ReadOffsetTable reads count uint32 offsets starting at the given position.
func ReadOffsetTable(data []byte, start int, count int) ([]uint32, error) {
	offsets := make([]uint32, count)
	for i := 0; i < count; i++ {
		off, err := binutil.U32LE(data, start+i*4)
		if err != nil {
			return nil, fmt.Errorf("reading offset %d: %w", i, err)
		}
		offsets[i] = off
	}
	return offsets, nil
}
