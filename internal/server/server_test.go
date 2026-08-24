package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelkhan/marky/internal/server"
)

func setup(t *testing.T, yamlContent, mdContent string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "doc.yaml")
	mdPath := filepath.Join(dir, "doc.md")
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	_ = os.WriteFile(mdPath, []byte(mdContent), 0644)
	return yamlPath, mdPath
}

const testYAML = "title: \"Test\"\nsections:\n  - type: paragraph\n    text: \"Hello\"\n"
const testMD = "# Test\n\nHello\n"

func TestFindFreePort(t *testing.T) {
	port, err := server.FindFreePort()
	if err != nil {
		t.Fatalf("FindFreePort: %v", err)
	}
	if port < 1024 || port > 65535 {
		t.Errorf("port %d out of expected range", port)
	}
}

func TestServer_ContentEndpoint(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/content", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(body["content"], "# Test") {
		t.Errorf("content = %q, expected to contain '# Test'", body["content"])
	}
}

func TestServer_SaveEndpoint(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	newMD := "# Test\n\nUpdated paragraph.\n"
	payload, _ := json.Marshal(map[string]string{"content": newMD})
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	// Verify .md file updated
	data, _ := os.ReadFile(mdPath)
	if !strings.Contains(string(data), "Updated paragraph") {
		t.Errorf("md file not updated: %s", data)
	}

	// Verify .yaml file updated
	yamlData, _ := os.ReadFile(yamlPath)
	if !strings.Contains(string(yamlData), "Updated paragraph") {
		t.Errorf("yaml file not updated: %s", yamlData)
	}
}

func TestServer_SaveEndpoint_InvalidMarkdown(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	// Markdown without H1 title is invalid for our schema
	payload, _ := json.Marshal(map[string]string{"content": "## No title here\n"})
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("expected non-200 for invalid markdown, got 200")
	}
}

func TestServer_SaveEndpoint_AtomicWrite(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	newMD := "# Test\n\nAtomic write check.\n"
	payload, _ := json.Marshal(map[string]string{"content": newMD})
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(httptest.NewRecorder(), req)

	// File must exist and be readable immediately after save
	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("md file unreadable after save: %v", err)
	}
	if len(data) == 0 {
		t.Error("md file is empty after save")
	}
}

func TestServer_EventsOnSave(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	newMD := "# Test\n\nEvent check.\n"
	payload, _ := json.Marshal(map[string]string{"content": newMD})
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(httptest.NewRecorder(), req)

	select {
	case ev := <-srv.Events():
		if ev.Status != "saved" {
			t.Errorf("event status = %q, want %q", ev.Status, "saved")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for save event")
	}
}

func TestServer_IndexServesHTML(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "marky editor") {
		t.Error("expected HTML page with 'marky editor'")
	}
}

func TestServer_StartAndStop(t *testing.T) {
	yamlPath, mdPath := setup(t, testYAML, testMD)
	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- srv.Start(ctx) }()

	// Wait for ready event
	select {
	case ev := <-srv.Events():
		if ev.Status != "ready" {
			t.Errorf("first event = %q, want ready", ev.Status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for server ready")
	}

	// Verify server is reachable
	resp, err := http.Get("http://localhost:" + itoa(srv.Port()) + "/content")
	if err != nil {
		t.Fatalf("server not reachable: %v", err)
	}
	resp.Body.Close()

	cancel()

	select {
	case err := <-errc:
		if err != nil && !strings.Contains(err.Error(), "Server closed") {
			t.Errorf("unexpected server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop after context cancel")
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
