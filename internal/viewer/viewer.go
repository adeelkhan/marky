package viewer

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/adeelkhan/marky/internal/server"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// ServerEventMsg wraps server.ServerEvent as a Bubbletea message.
type ServerEventMsg server.ServerEvent

var (
	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#e94560")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e94560")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50fa7b"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a0a0b0"))
)

type model struct {
	viewport viewport.Model
	content  string
	filePath string
	port     int
	events   <-chan server.ServerEvent
	status   string
	width    int
	height   int
	ready    bool
	showHelp bool
}

// New creates a new TUI model. Call tea.NewProgram(viewer.New(...)).Run() to start.
func New(content, filePath string, port int, events <-chan server.ServerEvent) tea.Model {
	return &model{
		content:  content,
		filePath: filePath,
		port:     port,
		events:   events,
		status:   "starting…",
	}
}

func waitForEvent(ch <-chan server.ServerEvent) tea.Cmd {
	return func() tea.Msg {
		return ServerEventMsg(<-ch)
	}
}

func (m *model) Init() tea.Cmd {
	return waitForEvent(m.events)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "o":
			cmds = append(cmds, openBrowser(fmt.Sprintf("http://localhost:%d", m.port)))
		case "?":
			m.showHelp = !m.showHelp
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		const fixedLines = 3 // header + 2 footer lines
		vpHeight := m.height - fixedLines
		if vpHeight < 1 {
			vpHeight = 1
		}
		if !m.ready {
			m.viewport = viewport.New(m.width, vpHeight)
			m.viewport.SetContent(m.renderMarkdown())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = vpHeight
		}

	case ServerEventMsg:
		ev := server.ServerEvent(msg)
		m.status = ev.Status
		if ev.Content != "" {
			m.content = ev.Content
			if m.ready {
				m.viewport.SetContent(m.renderMarkdown())
			}
		}
		cmds = append(cmds, waitForEvent(m.events))
	}

	var vpCmd tea.Cmd
	if m.ready {
		m.viewport, vpCmd = m.viewport.Update(msg)
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if !m.ready {
		return "Loading…\n"
	}

	header := headerStyle.Width(m.width).Render(
		fmt.Sprintf(" marky  ·  %s", m.filePath),
	)

	var footer1 string
	if m.showHelp {
		footer1 = footerStyle.Render("  HELP: ↑↓/jk scroll  o open browser  ? hide help  q quit")
	} else {
		footer1 = footerStyle.Render("  ↑↓ scroll  o open browser  ? help  q quit")
	}
	url := urlStyle.Render(fmt.Sprintf("http://localhost:%d", m.port))
	st := statusStyle.Render(fmt.Sprintf("[%s]", m.status))
	footer2 := footerStyle.Render(fmt.Sprintf("  %s  %s", url, st))

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, m.viewport.View(), footer1, footer2)
}

func (m *model) renderMarkdown() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		return m.content
	}
	out, err := r.Render(m.content)
	if err != nil {
		return m.content
	}
	return out
}

func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			cmd = exec.Command("cmd", "/c", "start", url)
		}
		_ = cmd.Start()
		return nil
	}
}
