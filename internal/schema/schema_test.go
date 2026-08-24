package schema_test

import (
	"fmt"
	"testing"

	"github.com/adeelkhan/marky/internal/schema"
)

func TestParse_ValidDocument(t *testing.T) {
	data := []byte(`
title: "Test Doc"
sections:
  - type: heading
    level: 2
    text: "Hello"
  - type: paragraph
    text: "World"
`)
	doc, err := schema.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "Test Doc" {
		t.Errorf("title = %q, want %q", doc.Title, "Test Doc")
	}
	if len(doc.Sections) != 2 {
		t.Fatalf("sections count = %d, want 2", len(doc.Sections))
	}
	if doc.Sections[0].Type != "heading" || doc.Sections[0].Level != 2 {
		t.Errorf("section[0] = %+v, want heading level 2", doc.Sections[0])
	}
}

func TestParse_MissingTitle(t *testing.T) {
	_, err := schema.Parse([]byte(`sections: []`))
	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}
}

func TestParse_UnknownSectionType(t *testing.T) {
	data := []byte("title: \"Doc\"\nsections:\n  - type: table\n    text: \"x\"\n")
	_, err := schema.Parse(data)
	if err == nil {
		t.Fatal("expected error for unknown section type, got nil")
	}
}

func TestParse_HeadingLevelOutOfRange(t *testing.T) {
	for _, level := range []int{0, 7} {
		data := []byte(fmt.Sprintf("title: \"Doc\"\nsections:\n  - type: heading\n    level: %d\n    text: \"Hi\"\n", level))
		_, err := schema.Parse(data)
		if err == nil {
			t.Errorf("expected error for heading level %d, got nil", level)
		}
	}
}

func TestParse_EmptyListItems(t *testing.T) {
	data := []byte("title: \"Doc\"\nsections:\n  - type: list\n    ordered: false\n    items: []\n")
	_, err := schema.Parse(data)
	if err == nil {
		t.Fatal("expected error for empty list items, got nil")
	}
}

func TestParse_LinkMissingURL(t *testing.T) {
	data := []byte("title: \"Doc\"\nsections:\n  - type: link\n    text: \"Click\"\n")
	_, err := schema.Parse(data)
	if err == nil {
		t.Fatal("expected error for link missing url, got nil")
	}
}

func TestParse_EmptySections(t *testing.T) {
	doc, err := schema.Parse([]byte("title: \"Doc\"\nsections: []\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Sections) != 0 {
		t.Errorf("want 0 sections, got %d", len(doc.Sections))
	}
}

func TestMarshal_RoundTrip(t *testing.T) {
	data := []byte("title: Test\nsections:\n  - type: paragraph\n    text: Hello\n")
	doc, err := schema.Parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := doc.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	doc2, err := schema.Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if doc2.Title != doc.Title || len(doc2.Sections) != len(doc.Sections) {
		t.Errorf("round-trip mismatch: got %+v, want %+v", doc2, doc)
	}
}
