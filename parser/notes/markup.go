package notes

import (
	"strings"

	"github.com/kevin-cantwell/reunion-explore/model"
)

const (
	openTag  = "\u00AB" // «
	closeTag = "\u00BB" // »
)

// ParseMarkup tokenizes and parses «»-delimited markup into a tree of MarkupNodes.
func ParseMarkup(text string) []model.MarkupNode {
	p := &markupParser{input: text, maxDepth: 20}
	return p.parse()
}

type markupParser struct {
	input    string
	pos      int
	maxDepth int
}

func (p *markupParser) parse() []model.MarkupNode {
	return p.parseUntil("", 0)
}

func (p *markupParser) parseUntil(endTag string, depth int) []model.MarkupNode {
	if depth > p.maxDepth {
		// Prevent infinite recursion
		return nil
	}

	var nodes []model.MarkupNode

	for p.pos < len(p.input) {
		// Find the next «
		nextOpen := strings.Index(p.input[p.pos:], openTag)
		if nextOpen == -1 {
			// No more tags - rest is plain text
			rest := p.input[p.pos:]
			if len(rest) > 0 {
				nodes = append(nodes, model.MarkupNode{Type: model.MarkupText, Text: rest})
			}
			p.pos = len(p.input)
			break
		}

		// Emit any text before the tag
		if nextOpen > 0 {
			nodes = append(nodes, model.MarkupNode{
				Type: model.MarkupText,
				Text: p.input[p.pos : p.pos+nextOpen],
			})
		}
		p.pos += nextOpen

		// Find the closing »
		tagStart := p.pos + len(openTag)
		nextClose := strings.Index(p.input[tagStart:], closeTag)
		if nextClose == -1 {
			// Malformed tag - emit rest as text
			nodes = append(nodes, model.MarkupNode{
				Type: model.MarkupText,
				Text: p.input[p.pos:],
			})
			p.pos = len(p.input)
			break
		}

		tagContent := p.input[tagStart : tagStart+nextClose]
		p.pos = tagStart + nextClose + len(closeTag)

		// Check if this is a closing tag
		if strings.HasPrefix(tagContent, "/") {
			closedName := tagContent[1:]
			if endTag != "" && closedName == endTag {
				// Found our matching close tag
				return nodes
			}
			// Mismatched close tag - emit as text and return to parent
			// (this handles malformed markup gracefully)
			continue
		}

		// Opening tag - parse its content
		node := p.parseTag(tagContent, depth)
		if node != nil {
			nodes = append(nodes, *node)
		}
	}

	return nodes
}

func (p *markupParser) parseTag(tag string, depth int) *model.MarkupNode {
	switch {
	case tag == "b":
		children := p.parseUntil("b", depth+1)
		return &model.MarkupNode{Type: model.MarkupBold, Children: children}
	case tag == "i":
		children := p.parseUntil("i", depth+1)
		return &model.MarkupNode{Type: model.MarkupItalic, Children: children}
	case tag == "u":
		children := p.parseUntil("u", depth+1)
		return &model.MarkupNode{Type: model.MarkupUnderline, Children: children}
	case strings.HasPrefix(tag, "ff="):
		value := strings.TrimPrefix(tag, "ff=")
		children := p.parseUntil("ff", depth+1)
		return &model.MarkupNode{Type: model.MarkupFontFlag, Value: value, Children: children}
	case strings.HasPrefix(tag, "c="):
		value := strings.TrimPrefix(tag, "c=")
		children := p.parseUntil("c", depth+1)
		return &model.MarkupNode{Type: model.MarkupColor, Value: value, Children: children}
	case strings.HasPrefix(tag, "url="):
		value := strings.TrimPrefix(tag, "url=")
		children := p.parseUntil("url", depth+1)
		return &model.MarkupNode{Type: model.MarkupURL, Value: value, Children: children}
	case strings.HasPrefix(tag, "s="):
		value := strings.TrimPrefix(tag, "s=")
		children := p.parseUntil("s", depth+1)
		return &model.MarkupNode{Type: model.MarkupSourceCitation, Value: value, Children: children}
	default:
		// Unknown tag, treat as text
		return &model.MarkupNode{
			Type: model.MarkupText,
			Text: openTag + tag + closeTag,
		}
	}
}
