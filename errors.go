package reunion

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrNotABundle     = errors.New("not a valid Reunion bundle directory")
	ErrMissingFile    = errors.New("required file missing from bundle")
	ErrBadMagic       = errors.New("unexpected file magic bytes")
	ErrUnsupportedVer = errors.New("unsupported Reunion version")
	ErrCorruptRecord  = errors.New("corrupt or unreadable record")
)

// ParseError represents a non-fatal error encountered during parsing.
type ParseError struct {
	File    string // source file within the bundle
	Offset  int    // byte offset if applicable, -1 otherwise
	Message string
	Err     error // underlying error, if any
}

func (e *ParseError) Error() string {
	var b strings.Builder
	if e.File != "" {
		b.WriteString(e.File)
		if e.Offset >= 0 {
			fmt.Fprintf(&b, "@0x%X", e.Offset)
		}
		b.WriteString(": ")
	}
	b.WriteString(e.Message)
	if e.Err != nil {
		fmt.Fprintf(&b, ": %v", e.Err)
	}
	return b.String()
}

func (e *ParseError) Unwrap() error { return e.Err }

// ErrorCollector accumulates non-fatal parse errors in a thread-safe manner.
type ErrorCollector struct {
	mu        sync.Mutex
	errors    []ParseError
	maxErrors int
}

// NewErrorCollector creates a collector that stops accepting errors after max.
// If max <= 0, there is no limit.
func NewErrorCollector(max int) *ErrorCollector {
	return &ErrorCollector{maxErrors: max}
}

// Add records a non-fatal error. Returns true if the error was added,
// false if the maximum has been reached.
func (c *ErrorCollector) Add(file string, offset int, msg string, err error) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxErrors > 0 && len(c.errors) >= c.maxErrors {
		return false
	}
	c.errors = append(c.errors, ParseError{
		File:    file,
		Offset:  offset,
		Message: msg,
		Err:     err,
	})
	return true
}

// Errors returns a copy of all collected errors.
func (c *ErrorCollector) Errors() []ParseError {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ParseError, len(c.errors))
	copy(out, c.errors)
	return out
}

// Len returns the number of collected errors.
func (c *ErrorCollector) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.errors)
}

// Full returns true if the collector has reached its maximum.
func (c *ErrorCollector) Full() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxErrors > 0 && len(c.errors) >= c.maxErrors
}
