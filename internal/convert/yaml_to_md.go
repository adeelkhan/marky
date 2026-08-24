package convert

import (
	"fmt"
	"strings"

	"github.com/adeelkhan/marky/internal/schema"
)

func YAMLToMarkdown(doc *schema.Document) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n", doc.Title))

	if doc.Description != "" {
		sb.WriteString(fmt.Sprintf("\n*%s*\n", doc.Description))
	}

	for _, s := range doc.Sections {
		sb.WriteString("\n")
		sb.WriteString(sectionToMarkdown(s))
		sb.WriteString("\n")
	}

	return sb.String()
}

func sectionToMarkdown(s schema.Section) string {
	switch s.Type {
	case "heading":
		return fmt.Sprintf("%s %s", strings.Repeat("#", s.Level), s.Text)
	case "paragraph":
		return s.Text
	case "code":
		body := strings.TrimRight(s.Body, "\n")
		return fmt.Sprintf("```%s\n%s\n```", s.Lang, body)
	case "list":
		lines := make([]string, len(s.Items))
		for i, item := range s.Items {
			if s.Ordered {
				lines[i] = fmt.Sprintf("%d. %s", i+1, item)
			} else {
				lines[i] = fmt.Sprintf("- %s", item)
			}
		}
		return strings.Join(lines, "\n")
	case "blockquote":
		return fmt.Sprintf("> %s", s.Text)
	case "rule":
		return "---"
	case "link":
		return fmt.Sprintf("[%s](%s)", s.Text, s.URL)
	default:
		return ""
	}
}
