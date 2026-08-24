package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeelkhan/marky/internal/convert"
	"github.com/adeelkhan/marky/internal/schema"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var inputDir, outputDir string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Convert YAML files to Markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, inputDir, outputDir)
		},
	}
	cmd.Flags().StringVar(&inputDir, "input", "./yamls", "Input directory containing .yaml files")
	cmd.Flags().StringVar(&outputDir, "output", "./", "Output directory for .md files")
	return cmd
}

func runGenerate(cmd *cobra.Command, inputDir, outputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("read input directory %q: %w", inputDir, err)
	}

	generated := 0
	failed := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		yamlPath := filepath.Join(inputDir, entry.Name())
		mdName := strings.TrimSuffix(entry.Name(), ".yaml") + ".md"
		mdPath := filepath.Join(outputDir, mdName)

		if err := GenerateOne(yamlPath, mdPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %s: %v\n", entry.Name(), err)
			failed++
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated: %s → %s\n", yamlPath, mdPath)
		generated++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n%d file(s) generated.\n", generated)
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed to generate", failed)
	}
	return nil
}

// GenerateOne converts a single YAML file to a Markdown file.
// Exported so the view command can auto-generate before launching the TUI.
func GenerateOne(yamlPath, mdPath string) error {
	doc, err := schema.ParseFile(yamlPath)
	if err != nil {
		return err
	}
	md := convert.YAMLToMarkdown(doc)
	return os.WriteFile(mdPath, []byte(md), 0644)
}
