package model

import "strings"

// MarkupNodeType classifies a node in parsed «» markup.
type MarkupNodeType int

const (
	MarkupText      MarkupNodeType = iota // plain text
	MarkupBold                            // «b»...«/b»
	MarkupItalic                          // «i»...«/i»
	MarkupUnderline                       // «u»...«/u»
	MarkupFontFlag                        // «ff=N»...«/ff»
	MarkupColor                           // «c=RRGGBBAA»...«/c»
	MarkupURL                             // «url=URL»...«/url»
	MarkupSourceCitation                  // «s=N»...«/s»
)

// MarkupNode represents a node in the parsed markup tree.
type MarkupNode struct {
	Type     MarkupNodeType `json:"type"`
	Text     string         `json:"text,omitempty"`     // for MarkupText
	Value    string         `json:"value,omitempty"`    // attribute value (font flag number, color, URL)
	Children []MarkupNode   `json:"children,omitempty"` // for container nodes
}

// PlainText recursively extracts text content from markup nodes,
// stripping all formatting tags and returning readable plain text.
func PlainText(nodes []MarkupNode) string {
	var b []byte
	for _, n := range nodes {
		if n.Type == MarkupText {
			b = append(b, n.Text...)
		} else {
			b = append(b, PlainText(n.Children)...)
		}
	}
	s := string(b)
	// Strip any remaining «...» markup sequences (unrecognized tags or
	// incomplete tags left from truncated data).
	for {
		i := strings.Index(s, "\u00AB")
		if i == -1 {
			break
		}
		j := strings.Index(s[i:], "\u00BB")
		if j == -1 {
			// Trailing incomplete tag — remove from « onward
			s = s[:i]
			break
		}
		s = s[:i] + s[i+j+len("\u00BB"):]
	}
	return s
}
