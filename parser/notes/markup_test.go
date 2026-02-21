package notes

import (
	"testing"

	"github.com/kedoco/reunion-explore/model"
)

func TestParseMarkup_PlainText(t *testing.T) {
	nodes := ParseMarkup("Hello, world!")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupText {
		t.Errorf("expected MarkupText, got %d", nodes[0].Type)
	}
	if nodes[0].Text != "Hello, world!" {
		t.Errorf("text = %q, want %q", nodes[0].Text, "Hello, world!")
	}
}

func TestParseMarkup_Bold(t *testing.T) {
	nodes := ParseMarkup("before\u00ABb\u00BBbold text\u00AB/b\u00BBafter")
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupText || nodes[0].Text != "before" {
		t.Errorf("node[0] = %+v, want text 'before'", nodes[0])
	}
	if nodes[1].Type != model.MarkupBold {
		t.Errorf("node[1] type = %d, want MarkupBold", nodes[1].Type)
	}
	if len(nodes[1].Children) != 1 || nodes[1].Children[0].Text != "bold text" {
		t.Errorf("bold children = %+v, want 'bold text'", nodes[1].Children)
	}
	if nodes[2].Type != model.MarkupText || nodes[2].Text != "after" {
		t.Errorf("node[2] = %+v, want text 'after'", nodes[2])
	}
}

func TestParseMarkup_Italic(t *testing.T) {
	nodes := ParseMarkup("\u00ABi\u00BBitalic\u00AB/i\u00BB")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupItalic {
		t.Errorf("type = %d, want MarkupItalic", nodes[0].Type)
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].Text != "italic" {
		t.Errorf("children = %+v, want 'italic'", nodes[0].Children)
	}
}

func TestParseMarkup_FontFlag(t *testing.T) {
	nodes := ParseMarkup("\u00ABff=1\u00BBstyled\u00AB/ff\u00BB")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupFontFlag {
		t.Errorf("type = %d, want MarkupFontFlag", nodes[0].Type)
	}
	if nodes[0].Value != "1" {
		t.Errorf("value = %q, want %q", nodes[0].Value, "1")
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].Text != "styled" {
		t.Errorf("children = %+v, want 'styled'", nodes[0].Children)
	}
}

func TestParseMarkup_Nested(t *testing.T) {
	// «b»bold «i»bold-italic«/i»«/b»
	input := "\u00ABb\u00BBbold \u00ABi\u00BBbold-italic\u00AB/i\u00BB\u00AB/b\u00BB"
	nodes := ParseMarkup(input)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	bold := nodes[0]
	if bold.Type != model.MarkupBold {
		t.Fatalf("expected MarkupBold, got %d", bold.Type)
	}
	if len(bold.Children) != 2 {
		t.Fatalf("bold children count = %d, want 2", len(bold.Children))
	}
	if bold.Children[0].Text != "bold " {
		t.Errorf("first child text = %q, want %q", bold.Children[0].Text, "bold ")
	}
	if bold.Children[1].Type != model.MarkupItalic {
		t.Errorf("second child type = %d, want MarkupItalic", bold.Children[1].Type)
	}
	if len(bold.Children[1].Children) != 1 || bold.Children[1].Children[0].Text != "bold-italic" {
		t.Errorf("italic children = %+v", bold.Children[1].Children)
	}
}

func TestParseMarkup_URL(t *testing.T) {
	nodes := ParseMarkup("\u00ABurl=https://example.com\u00BBclick\u00AB/url\u00BB")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupURL {
		t.Errorf("type = %d, want MarkupURL", nodes[0].Type)
	}
	if nodes[0].Value != "https://example.com" {
		t.Errorf("value = %q, want URL", nodes[0].Value)
	}
}

func TestParseMarkup_Color(t *testing.T) {
	nodes := ParseMarkup("\u00ABc=FF0000FF\u00BBred\u00AB/c\u00BB")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupColor {
		t.Errorf("type = %d, want MarkupColor", nodes[0].Type)
	}
	if nodes[0].Value != "FF0000FF" {
		t.Errorf("value = %q, want %q", nodes[0].Value, "FF0000FF")
	}
}

func TestParseMarkup_SourceCitation(t *testing.T) {
	// «s=7»cited text«/s»
	nodes := ParseMarkup("before\u00ABs=7\u00BBcited text\u00AB/s\u00BBafter")
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if nodes[0].Type != model.MarkupText || nodes[0].Text != "before" {
		t.Errorf("node[0] = %+v, want text 'before'", nodes[0])
	}
	if nodes[1].Type != model.MarkupSourceCitation {
		t.Errorf("node[1] type = %d, want MarkupSourceCitation", nodes[1].Type)
	}
	if nodes[1].Value != "7" {
		t.Errorf("node[1] value = %q, want %q", nodes[1].Value, "7")
	}
	if len(nodes[1].Children) != 1 || nodes[1].Children[0].Text != "cited text" {
		t.Errorf("source citation children = %+v, want 'cited text'", nodes[1].Children)
	}
	if nodes[2].Type != model.MarkupText || nodes[2].Text != "after" {
		t.Errorf("node[2] = %+v, want text 'after'", nodes[2])
	}
}

func TestPlainText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "Hello", "Hello"},
		{"bold", "\u00ABb\u00BBbold\u00AB/b\u00BB", "bold"},
		{"nested", "\u00ABb\u00BBbold \u00ABi\u00BBitalic\u00AB/i\u00BB\u00AB/b\u00BB", "bold italic"},
		{"font flag", "\u00ABff=1\u00BBtext\u00AB/ff\u00BB", "text"},
		{"mixed", "before \u00ABb\u00BBmiddle\u00AB/b\u00BB after", "before middle after"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := ParseMarkup(tt.input)
			got := model.PlainText(nodes)
			if got != tt.want {
				t.Errorf("PlainText() = %q, want %q", got, tt.want)
			}
		})
	}
}
