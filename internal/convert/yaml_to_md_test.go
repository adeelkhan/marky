// internal/convert/yaml_to_md_test.go
package convert_test

import (
	"strings"
	"testing"

	"github.com/adeelkhan/marky/internal/convert"
	"github.com/adeelkhan/marky/internal/schema"
)

func doc(sections ...schema.Section) *schema.Document {
	return &schema.Document{Title: "Test", Sections: sections}
}

func TestYAMLToMarkdown_TitleOnly(t *testing.T) {
	out := convert.YAMLToMarkdown(&schema.Document{Title: "Hello"})
	if !strings.Contains(out, "# Hello") {
		t.Errorf("expected '# Hello' in output, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Description(t *testing.T) {
	d := &schema.Document{Title: "T", Description: "A subtitle"}
	out := convert.YAMLToMarkdown(d)
	if !strings.Contains(out, "*A subtitle*") {
		t.Errorf("expected italic description, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Heading(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "heading", Level: 3, Text: "Intro"}))
	if !strings.Contains(out, "### Intro") {
		t.Errorf("expected '### Intro', got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Paragraph(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "paragraph", Text: "Hello **world**"}))
	if !strings.Contains(out, "Hello **world**") {
		t.Errorf("expected paragraph text, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_CodeBlock(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "code", Lang: "go", Body: "fmt.Println()\n"}))
	if !strings.Contains(out, "```go") || !strings.Contains(out, "fmt.Println()") {
		t.Errorf("expected fenced code block, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_UnorderedList(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "list", Ordered: false, Items: []string{"a", "b"}}))
	if !strings.Contains(out, "- a") || !strings.Contains(out, "- b") {
		t.Errorf("expected unordered list, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_OrderedList(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "list", Ordered: true, Items: []string{"first", "second"}}))
	if !strings.Contains(out, "1. first") || !strings.Contains(out, "2. second") {
		t.Errorf("expected ordered list, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Blockquote(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "blockquote", Text: "Quote me"}))
	if !strings.Contains(out, "> Quote me") {
		t.Errorf("expected blockquote, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Rule(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "rule"}))
	if !strings.Contains(out, "---") {
		t.Errorf("expected horizontal rule, got:\n%s", out)
	}
}

func TestYAMLToMarkdown_Link(t *testing.T) {
	out := convert.YAMLToMarkdown(doc(schema.Section{Type: "link", Text: "Click", URL: "https://example.com"}))
	if !strings.Contains(out, "[Click](https://example.com)") {
		t.Errorf("expected markdown link, got:\n%s", out)
	}
}
