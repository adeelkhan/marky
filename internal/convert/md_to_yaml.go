package convert

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/adeelkhan/marky/internal/schema"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func MarkdownToYAML(src []byte) (*schema.Document, error) {
	md := goldmark.New()
	reader := text.NewReader(src)
	node := md.Parser().Parse(reader)

	var result schema.Document

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only handle direct children of the document root
		if n.Parent() == nil || n.Parent().Kind() != ast.KindDocument {
			if n.Kind() == ast.KindDocument {
				return ast.WalkContinue, nil
			}
			return ast.WalkSkipChildren, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			txt := extractRawLines(node, src)
			if node.Level == 1 && result.Title == "" {
				result.Title = txt
			} else {
				result.Sections = append(result.Sections, schema.Section{
					Type:  "heading",
					Level: node.Level,
					Text:  txt,
				})
			}
			return ast.WalkSkipChildren, nil

		case *ast.Paragraph:
			if info, ok := soloLink(node, src); ok {
				result.Sections = append(result.Sections, schema.Section{
					Type: "link",
					Text: info[0],
					URL:  info[1],
				})
			} else if isEmphasisOnly(node) && result.Title != "" && result.Description == "" && len(result.Sections) == 0 {
				result.Description = extractRawLines(node, src)
				result.Description = strings.Trim(result.Description, "*_")
			} else {
				result.Sections = append(result.Sections, schema.Section{
					Type: "paragraph",
					Text: extractRawLines(node, src),
				})
			}
			return ast.WalkSkipChildren, nil

		case *ast.FencedCodeBlock:
			lang := string(node.Language(src))
			var body bytes.Buffer
			for i := 0; i < node.Lines().Len(); i++ {
				line := node.Lines().At(i)
				body.Write(line.Value(src))
			}
			result.Sections = append(result.Sections, schema.Section{
				Type: "code",
				Lang: lang,
				Body: body.String(),
			})
			return ast.WalkSkipChildren, nil

		case *ast.List:
			var items []string
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				items = append(items, extractListItem(child, src))
			}
			result.Sections = append(result.Sections, schema.Section{
				Type:    "list",
				Ordered: node.IsOrdered(),
				Items:   items,
			})
			return ast.WalkSkipChildren, nil

		case *ast.Blockquote:
			var parts []string
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				if p, ok := child.(*ast.Paragraph); ok {
					parts = append(parts, extractRawLines(p, src))
				}
			}
			result.Sections = append(result.Sections, schema.Section{
				Type: "blockquote",
				Text: strings.Join(parts, " "),
			})
			return ast.WalkSkipChildren, nil

		case *ast.ThematicBreak:
			result.Sections = append(result.Sections, schema.Section{Type: "rule"})
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk ast: %w", err)
	}
	if result.Title == "" {
		return nil, fmt.Errorf("document has no title (missing H1 heading)")
	}
	return &result, nil
}

// soloLink returns [text, url] if the paragraph contains only a single link.
func soloLink(n *ast.Paragraph, src []byte) ([2]string, bool) {
	if n.ChildCount() != 1 {
		return [2]string{}, false
	}
	link, ok := n.FirstChild().(*ast.Link)
	if !ok {
		return [2]string{}, false
	}
	var sb strings.Builder
	for child := link.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			sb.Write(t.Segment.Value(src))
		}
	}
	return [2]string{strings.TrimSpace(sb.String()), string(link.Destination)}, true
}

func isEmphasisOnly(n *ast.Paragraph) bool {
	return n.ChildCount() == 1 && n.FirstChild().Kind() == ast.KindEmphasis
}

// extractRawLines collects the raw source lines belonging to node.
func extractRawLines(n ast.Node, src []byte) string {
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		// Fall back to walking text children (for headings)
		var sb strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				sb.Write(t.Segment.Value(src))
			} else {
				// Inline element — collect its text children
				for gc := child.FirstChild(); gc != nil; gc = gc.NextSibling() {
					if t, ok := gc.(*ast.Text); ok {
						sb.Write(t.Segment.Value(src))
					}
				}
			}
		}
		return strings.TrimSpace(sb.String())
	}
	var sb strings.Builder
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		sb.Write(line.Value(src))
	}
	return strings.TrimSpace(sb.String())
}

func extractListItem(n ast.Node, src []byte) string {
	var sb strings.Builder
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if p, ok := child.(*ast.Paragraph); ok {
			sb.WriteString(extractRawLines(p, src))
		} else if tb, ok := child.(*ast.TextBlock); ok {
			sb.WriteString(extractRawLines(tb, src))
		}
	}
	return sb.String()
}
