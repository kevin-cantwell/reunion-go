package member

import (
	"github.com/kevin-cantwell/reunion-explore/bundle"
	"github.com/kevin-cantwell/reunion-explore/model"
	"github.com/kevin-cantwell/reunion-explore/parser/changes"
)

// ParseMembers creates Member models from bundle member directories.
func ParseMembers(members []bundle.MemberDir) ([]model.Member, error) {
	var result []model.Member

	for _, md := range members {
		m := model.Member{
			Name:      md.Name,
			DirPath:   md.Path,
			NoteFiles: md.NoteFiles,
			HasMedia:  md.MediaDir != "",
		}

		if md.Changes != "" {
			m.HasChanges = true
			recs, err := changes.ParseChanges(md.Changes)
			if err == nil {
				m.Changes = recs
			}
		}

		result = append(result, m)
	}

	return result, nil
}
