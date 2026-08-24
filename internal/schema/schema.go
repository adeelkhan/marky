package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Author      string    `yaml:"author,omitempty"`
	Date        string    `yaml:"date,omitempty"`
	Sections    []Section `yaml:"sections"`
}

type Section struct {
	Type    string   `yaml:"type"`
	Level   int      `yaml:"level,omitempty"`
	Text    string   `yaml:"text,omitempty"`
	Lang    string   `yaml:"lang,omitempty"`
	Body    string   `yaml:"body,omitempty"`
	Ordered bool     `yaml:"ordered,omitempty"`
	Items   []string `yaml:"items,omitempty"`
	URL     string   `yaml:"url,omitempty"`
}

var validTypes = map[string]bool{
	"heading": true, "paragraph": true, "code": true,
	"list": true, "blockquote": true, "rule": true, "link": true,
}

func Parse(data []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	return &doc, nil
}

func ParseFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}
	return Parse(data)
}

func (d *Document) Validate() error {
	if d.Title == "" {
		return fmt.Errorf("document title is required")
	}
	for i, s := range d.Sections {
		if err := s.validate(i); err != nil {
			return err
		}
	}
	return nil
}

func (s *Section) validate(idx int) error {
	if !validTypes[s.Type] {
		return fmt.Errorf("section %d: unknown type %q", idx, s.Type)
	}
	switch s.Type {
	case "heading":
		if s.Level < 1 || s.Level > 6 {
			return fmt.Errorf("section %d: heading level must be 1-6, got %d", idx, s.Level)
		}
	case "list":
		if len(s.Items) == 0 {
			return fmt.Errorf("section %d: list must have at least one item", idx)
		}
	case "link":
		if s.URL == "" {
			return fmt.Errorf("section %d: link url is required", idx)
		}
	}
	return nil
}

func (d *Document) Marshal() ([]byte, error) {
	return yaml.Marshal(d)
}
