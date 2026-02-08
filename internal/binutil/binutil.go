// Package binutil provides low-level binary reading utilities for parsing
// Reunion family file formats.
package binutil

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrShortRead = errors.New("short read")

// ReadU16LE reads a little-endian uint16 from r.
func ReadU16LE(r io.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf[:]), nil
}

// ReadU32LE reads a little-endian uint32 from r.
func ReadU32LE(r io.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// U16LE reads a little-endian uint16 from a byte slice at the given offset.
func U16LE(data []byte, offset int) (uint16, error) {
	if offset+2 > len(data) {
		return 0, fmt.Errorf("%w: need 2 bytes at offset %d, have %d", ErrShortRead, offset, len(data))
	}
	return binary.LittleEndian.Uint16(data[offset:]), nil
}

// U32LE reads a little-endian uint32 from a byte slice at the given offset.
func U32LE(data []byte, offset int) (uint32, error) {
	if offset+4 > len(data) {
		return 0, fmt.Errorf("%w: need 4 bytes at offset %d, have %d", ErrShortRead, offset, len(data))
	}
	return binary.LittleEndian.Uint32(data[offset:]), nil
}

// ReadNullTermString reads a null-terminated string from a byte slice starting at offset.
// Returns the string and the number of bytes consumed (including the null terminator).
func ReadNullTermString(data []byte, offset int) (string, int) {
	for i := offset; i < len(data); i++ {
		if data[i] == 0 {
			return string(data[offset:i]), i - offset + 1
		}
	}
	return string(data[offset:]), len(data) - offset
}

// ReadLenPrefixedString reads a string prefixed by its length as a uint8.
func ReadLenPrefixedString(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", 0, fmt.Errorf("%w: no length byte at offset %d", ErrShortRead, offset)
	}
	length := int(data[offset])
	end := offset + 1 + length
	if end > len(data) {
		return "", 0, fmt.Errorf("%w: string length %d exceeds data at offset %d", ErrShortRead, length, offset)
	}
	return string(data[offset+1 : end]), 1 + length, nil
}

// ReadNewlineTermString reads a newline-terminated string from data starting at offset.
func ReadNewlineTermString(data []byte, offset int) (string, int) {
	for i := offset; i < len(data); i++ {
		if data[i] == '\n' {
			return string(data[offset:i]), i - offset + 1
		}
	}
	return string(data[offset:]), len(data) - offset
}
