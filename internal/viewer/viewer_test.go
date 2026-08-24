package viewer_test

import (
	"strings"
	"testing"

	"github.com/adeelkhan/marky/internal/server"
	"github.com/adeelkhan/marky/internal/viewer"
	tea "github.com/charmbracelet/bubbletea"
)

func newModel(content string) tea.Model {
	ch := make(chan server.ServerEvent, 1)
	return viewer.New(content, "test.yaml", 8080, ch)
}

func TestViewer_InitialState(t *testing.T) {
	m := newModel("# Hello\n")
	if m == nil {
		t.Fatal("viewer.New returned nil")
	}
}

func TestViewer_QuitKey(t *testing.T) {
	m := newModel("# Hello\n")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if updated == nil {
		t.Fatal("Update returned nil model")
	}
	if cmd == nil {
		t.Fatal("expected quit command from 'q' key, got nil")
	}
}

func TestViewer_WindowResize(t *testing.T) {
	m := newModel("# Hello\n")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if updated == nil {
		t.Fatal("Update after WindowSizeMsg returned nil")
	}
	// View should not panic after resize
	view := updated.View()
	if view == "" {
		t.Error("View() returned empty string after resize")
	}
}

func TestViewer_ServerStatusUpdate(t *testing.T) {
	ch := make(chan server.ServerEvent, 1)
	m := viewer.New("# Hello\n", "test.yaml", 8080, ch)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated2, _ := updated.Update(viewer.ServerEventMsg(server.ServerEvent{Status: "ready"}))
	view := updated2.View()
	if view == "" {
		t.Error("View() returned empty string after status update")
	}
	if !strings.Contains(view, "ready") {
		t.Errorf("expected 'ready' status in view, got: %s", view)
	}
}

func TestViewer_ContentUpdate(t *testing.T) {
	ch := make(chan server.ServerEvent, 1)
	m := viewer.New("# Hello\n", "test.yaml", 8080, ch)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	newContent := "# Updated\n\nNew paragraph.\n"
	updated2, _ := updated.Update(viewer.ServerEventMsg(server.ServerEvent{
		Status:  "saved",
		Content: newContent,
	}))
	view := updated2.View()
	if !strings.Contains(view, "Updated") {
		t.Errorf("expected updated content in view, got: %s", view)
	}
}
