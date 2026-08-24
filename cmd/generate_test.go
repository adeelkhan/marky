package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelkhan/marky/cmd"
)

func TestGenerateOne_CreatesMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "test.yaml")
	mdPath := filepath.Join(dir, "test.md")

	yamlContent := `title: "Hello World"
sections:
  - type: paragraph
    text: "This is a test."
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := cmd.GenerateOne(yamlPath, mdPath); err != nil {
		t.Fatalf("GenerateOne: %v", err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(data), "# Hello World") {
		t.Errorf("expected '# Hello World' in output, got:\n%s", data)
	}
	if !strings.Contains(string(data), "This is a test.") {
		t.Errorf("expected paragraph text in output, got:\n%s", data)
	}
}

func TestGenerateOne_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	_ = os.WriteFile(yamlPath, []byte("sections: []"), 0644) // missing title
	err := cmd.GenerateOne(yamlPath, filepath.Join(dir, "bad.md"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestGenerateOne_MissingFile(t *testing.T) {
	err := cmd.GenerateOne("/nonexistent/path.yaml", "/tmp/out.md")
	if err == nil {
		t.Fatal("expected error for missing input file, got nil")
	}
}

func TestGenerateCommand_ProcessesDirectory(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	for _, name := range []string{"a.yaml", "b.yaml"} {
		content := "title: \"Doc " + name + "\"\nsections: []\n"
		_ = os.WriteFile(filepath.Join(inputDir, name), []byte(content), 0644)
	}
	// Non-yaml file should be ignored
	_ = os.WriteFile(filepath.Join(inputDir, "notes.txt"), []byte("ignore me"), 0644)

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--input", inputDir, "--output", outputDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("generate command: %v", err)
	}

	for _, name := range []string{"a.md", "b.md"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outputDir, "notes.md")); err == nil {
		t.Error("notes.md should not have been created")
	}
}

func TestGenerateCommand_MissingInputDir(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--input", "/nonexistent/dir"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input dir, got nil")
	}
}
