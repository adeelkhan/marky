package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeelkhan/marky/internal/server"
	"github.com/adeelkhan/marky/internal/viewer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <file.yaml>",
		Short: "View a YAML document in the terminal with live browser editing",
		Args:  cobra.ExactArgs(1),
		RunE:  runView,
	}
}

func runView(cmd *cobra.Command, args []string) error {
	yamlPath := args[0]
	if _, err := os.Stat(yamlPath); err != nil {
		return fmt.Errorf("file not found: %s", yamlPath)
	}

	base := strings.TrimSuffix(filepath.Base(yamlPath), ".yaml")
	mdPath := filepath.Join(filepath.Dir(yamlPath), base+".md")

	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		if err := GenerateOne(yamlPath, mdPath); err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}
	}

	content, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("read markdown: %w", err)
	}

	srv, err := server.New(yamlPath, mdPath)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel() // stop the TUI if server fails to start
		}
	}()

	m := viewer.New(string(content), filepath.Base(yamlPath), srv.Port(), srv.Events())
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
