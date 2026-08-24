# marky

A CLI tool that treats structured YAML as the source of truth for Markdown documents. Write your content in YAML, generate polished Markdown, and edit it in a browser — changes sync back to both files automatically.

## Features

- **Generate** — convert a directory of YAML files to `.md` files in one command
- **View** — render any YAML document as formatted Markdown in an interactive terminal TUI
- **Edit** — press `o` in the TUI to open a split-pane browser editor; save syncs changes back to both `.md` and `.yaml`

## Install

```bash
git clone https://github.com/adeelkhan/marky
cd marky
make install
```

Or build locally:

```bash
make build
# binary at ./bin/marky
```

## Usage

### Generate Markdown from YAML

```bash
marky generate --input ./yamls --output ./
```

Converts every `*.yaml` file in `./yamls` to a corresponding `.md` file. Exits non-zero if any file fails.

| Flag | Default | Description |
|---|---|---|
| `--input` | `./yamls` | Directory containing `.yaml` source files |
| `--output` | `./` | Directory to write `.md` files into |

### View and edit a document

```bash
marky view yamls/example.yaml
```

Opens an interactive TUI showing the rendered document. If the corresponding `.md` file doesn't exist yet, it is generated automatically.

**Keybindings:**

| Key | Action |
|---|---|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `o` | Open browser editor |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

Pressing `o` opens `http://localhost:<port>` in your browser — a self-contained split-pane editor with live preview. `ctrl+s` or the Save button writes the changes back to both the `.md` and `.yaml` files atomically.

## YAML schema

Each file in the input directory represents one document:

```yaml
title: "My Document"        # required — becomes # H1
description: "A subtitle"   # optional
author: "Adeel"             # optional metadata (preserved through edits)
date: "2026-08-25"          # optional metadata (preserved through edits)
sections:
  - type: heading
    level: 2
    text: "Introduction"

  - type: paragraph
    text: "Body text with **bold** and *italic* inline."

  - type: code
    lang: go
    body: |
      fmt.Println("hello")

  - type: list
    ordered: false
    items:
      - "First item"
      - "Second item"

  - type: blockquote
    text: "A quoted passage."

  - type: rule          # horizontal rule ---

  - type: link
    text: "Visit site"
    url: "https://example.com"
```

**Supported section types:** `heading`, `paragraph`, `code`, `list`, `blockquote`, `rule`, `link`

Inline Markdown (`**bold**`, `*italic*`, `` `code` ``) is allowed inside `text` and list `items` fields and is passed through as-is.

## Development

```bash
make test        # run all tests
make coverage    # generate HTML coverage report
make build       # build to bin/marky
make lint        # run golangci-lint
make clean       # remove bin/ and coverage.out
```

**Project layout:**

```
cmd/              Cobra commands (generate, view)
internal/
  schema/         YAML document types and validation
  convert/        YAML→MD and MD→YAML engines
  server/         HTTP server and browser editor
  viewer/         Bubbletea TUI
yamls/            Example input files
```

## Requirements

- Go 1.22+
