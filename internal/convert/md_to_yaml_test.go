// internal/convert/md_to_yaml_test.go
package convert_test

import (
	"testing"

	"github.com/adeelkhan/marky/internal/convert"
	"github.com/adeelkhan/marky/internal/schema"
)

func roundTrip(t *testing.T, in *schema.Document) *schema.Document {
	t.Helper()
	md := convert.YAMLToMarkdown(in)
	out, err := convert.MarkdownToYAML([]byte(md))
	if err != nil {
		t.Fatalf("MarkdownToYAML: %v", err)
	}
	return out
}

func TestMarkdownToYAML_Title(t *testing.T) {
	src := []byte("# My Title\n")
	doc, err := convert.MarkdownToYAML(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "My Title" {
		t.Errorf("title = %q, want %q", doc.Title, "My Title")
	}
}

func TestMarkdownToYAML_NoTitle(t *testing.T) {
	_, err := convert.MarkdownToYAML([]byte("## Not a title\n"))
	if err == nil {
		t.Fatal("expected error for missing H1 title, got nil")
	}
}

func TestRoundTrip_Heading(t *testing.T) {
	in := doc(schema.Section{Type: "heading", Level: 3, Text: "Section"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "heading" || out.Sections[0].Level != 3 {
		t.Errorf("round-trip heading: got %+v", out.Sections)
	}
}

func TestRoundTrip_Paragraph(t *testing.T) {
	in := doc(schema.Section{Type: "paragraph", Text: "Hello **world**"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "paragraph" {
		t.Errorf("round-trip paragraph: got %+v", out.Sections)
	}
	if out.Sections[0].Text != "Hello **world**" {
		t.Errorf("paragraph text = %q, want %q", out.Sections[0].Text, "Hello **world**")
	}
}

func TestRoundTrip_CodeBlock(t *testing.T) {
	in := doc(schema.Section{Type: "code", Lang: "go", Body: "fmt.Println()\n"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "code" || out.Sections[0].Lang != "go" {
		t.Errorf("round-trip code: got %+v", out.Sections)
	}
}

func TestRoundTrip_UnorderedList(t *testing.T) {
	in := doc(schema.Section{Type: "list", Ordered: false, Items: []string{"a", "b", "c"}})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "list" || out.Sections[0].Ordered {
		t.Errorf("round-trip unordered list: got %+v", out.Sections)
	}
	if len(out.Sections[0].Items) != 3 {
		t.Errorf("list items count = %d, want 3", len(out.Sections[0].Items))
	}
}

func TestRoundTrip_OrderedList(t *testing.T) {
	in := doc(schema.Section{Type: "list", Ordered: true, Items: []string{"first", "second"}})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || !out.Sections[0].Ordered {
		t.Errorf("round-trip ordered list: got %+v", out.Sections)
	}
}

func TestRoundTrip_Blockquote(t *testing.T) {
	in := doc(schema.Section{Type: "blockquote", Text: "A quote"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "blockquote" {
		t.Errorf("round-trip blockquote: got %+v", out.Sections)
	}
}

func TestRoundTrip_Rule(t *testing.T) {
	in := doc(schema.Section{Type: "rule"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "rule" {
		t.Errorf("round-trip rule: got %+v", out.Sections)
	}
}

func TestRoundTrip_Link(t *testing.T) {
	in := doc(schema.Section{Type: "link", Text: "Click", URL: "https://example.com"})
	out := roundTrip(t, in)
	if len(out.Sections) != 1 || out.Sections[0].Type != "link" {
		t.Errorf("round-trip link: got %+v", out.Sections)
	}
	if out.Sections[0].URL != "https://example.com" {
		t.Errorf("link url = %q, want %q", out.Sections[0].URL, "https://example.com")
	}
}

func TestMarkdownToYAML_StandaloneLinkVsInline(t *testing.T) {
	// Standalone link paragraph → link section
	solo := []byte("# Title\n\n[Click](https://x.com)\n")
	doc, err := convert.MarkdownToYAML(solo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Sections[0].Type != "link" {
		t.Errorf("standalone link: got type %q, want link", doc.Sections[0].Type)
	}

	// Link among other text → paragraph section
	inline := []byte("# Title\n\nSee [Click](https://x.com) for more.\n")
	doc2, err := convert.MarkdownToYAML(inline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc2.Sections[0].Type != "paragraph" {
		t.Errorf("inline link: got type %q, want paragraph", doc2.Sections[0].Type)
	}
}
