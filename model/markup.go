package model

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
)

// MarkupNode represents a node in the parsed markup tree.
type MarkupNode struct {
	Type     MarkupNodeType `json:"type"`
	Text     string         `json:"text,omitempty"`     // for MarkupText
	Value    string         `json:"value,omitempty"`    // attribute value (font flag number, color, URL)
	Children []MarkupNode   `json:"children,omitempty"` // for container nodes
}
