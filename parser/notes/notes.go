package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/kevin-cantwell/reunion-go/model"
)

// noteFilePattern matches note filenames like "p1-1106-13.note"
var noteFilePattern = regexp.MustCompile(`^p(\d+)-(\d+)-(\d+)\.note$`)

// ParseNoteFile reads and parses a single .note file.
func ParseNoteFile(path string) (*model.Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading note file %s: %w", path, err)
	}

	text := string(data)
	base := filepath.Base(path)

	note := &model.Note{
		Filename: base,
		RawText:  text,
		Markup:   ParseMarkup(text),
	}

	// Extract person ID and event/source info from filename
	matches := noteFilePattern.FindStringSubmatch(base)
	if len(matches) == 4 {
		if pid, err := strconv.Atoi(matches[1]); err == nil {
			note.PersonID = pid
		}
		if tag, err := strconv.Atoi(matches[2]); err == nil {
			note.EventTag = tag
		}
		if src, err := strconv.Atoi(matches[3]); err == nil {
			note.SourceID = src
		}
	}

	return note, nil
}

// ParseAllNotes discovers and parses all note files from the given paths.
func ParseAllNotes(paths []string) ([]model.Note, error) {
	var notes []model.Note
	for _, path := range paths {
		note, err := ParseNoteFile(path)
		if err != nil {
			continue // skip unparseable notes
		}
		notes = append(notes, *note)
	}
	return notes, nil
}
