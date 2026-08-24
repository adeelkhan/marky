package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adeelkhan/marky/internal/convert"
	"github.com/adeelkhan/marky/internal/schema"
	"github.com/gorilla/websocket"
)

// ServerEvent is sent to the TUI when server state changes.
type ServerEvent struct {
	Status  string `json:"status"` // "ready", "saved", "error"
	Time    string `json:"time,omitempty"`
	Error   string `json:"error,omitempty"`
	Content string `json:"content,omitempty"`
}

// Server embeds the HTTP handler so it can be used with httptest.
type Server struct {
	yamlPath  string
	mdPath    string
	port      int
	events    chan ServerEvent
	mu        sync.RWMutex
	mdContent string
	clients   map[*websocket.Conn]bool
	clientMu  sync.Mutex
	upgrader  websocket.Upgrader
	mux       *http.ServeMux
}

// New creates a new Server. It finds a free port and reads the initial markdown content.
func New(yamlPath, mdPath string) (*Server, error) {
	port, err := FindFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	content, err := os.ReadFile(mdPath)
	if err != nil {
		return nil, fmt.Errorf("read markdown file: %w", err)
	}

	s := &Server{
		yamlPath:  yamlPath,
		mdPath:    mdPath,
		port:      port,
		events:    make(chan ServerEvent, 16),
		mdContent: string(content),
		clients:   make(map[*websocket.Conn]bool),
		upgrader:  websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/content", s.handleContent)
	s.mux.HandleFunc("/save", s.handleSave)
	s.mux.HandleFunc("/ws", s.handleWS)
	return s, nil
}

// Port returns the port the server is (or will be) listening on.
func (s *Server) Port() int { return s.port }

// Events returns a channel of server events.
func (s *Server) Events() <-chan ServerEvent { return s.events }

// ServeHTTP delegates to the internal mux, enabling use with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

// Start listens on s.port, sends a "ready" event, then serves until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", srv.Addr, err)
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	s.events <- ServerEvent{Status: "ready"}
	return srv.Serve(ln)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(editorHTML))
}

func (s *Server) handleContent(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	content := s.mdContent
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"content": content})
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	doc, err := convert.MarkdownToYAML([]byte(payload.Content))
	if err != nil {
		ev := ServerEvent{Status: "error", Error: err.Error()}
		s.broadcast(ev)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Carry forward metadata (author, date) from existing YAML that markdown cannot encode.
	if existing, err2 := schema.ParseFile(s.yamlPath); err2 == nil {
		doc.Author = existing.Author
		doc.Date = existing.Date
	}

	if err := writeAtomic(s.mdPath, []byte(payload.Content)); err != nil {
		http.Error(w, "write md failed", http.StatusInternalServerError)
		return
	}

	yamlData, err := doc.Marshal()
	if err != nil {
		http.Error(w, "marshal yaml failed", http.StatusInternalServerError)
		return
	}
	if err := writeAtomic(s.yamlPath, yamlData); err != nil {
		http.Error(w, "write yaml failed", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.mdContent = payload.Content
	s.mu.Unlock()

	ev := ServerEvent{
		Status:  "saved",
		Time:    time.Now().Format("15:04"),
		Content: payload.Content,
	}
	s.broadcast(ev)
	s.events <- ev

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.clientMu.Lock()
	s.clients[conn] = true
	s.clientMu.Unlock()

	defer func() {
		s.clientMu.Lock()
		delete(s.clients, conn)
		s.clientMu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *Server) broadcast(event ServerEvent) {
	data, _ := json.Marshal(event)
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	for conn := range s.clients {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".marky-tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// FindFreePort returns an available TCP port on localhost.
func FindFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}
