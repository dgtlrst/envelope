package app

import (
	"envelope/pkg/edit"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width, height int
	viewport      viewport.Model
	buffer        *edit.TextBuffer
	cursor        *edit.CursorPointer
	ready         bool
}

func NewModel() Model {
	return Model{
		buffer: edit.NewTextBuffer(),
		cursor: edit.NewCursor(0, 0),
		ready:  false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// insert new line (replace content with buf)
			m.buffer.InsertNewLine(m.cursor.Y, m.cursor.X)
			m.cursor.Y++
			m.cursor.X = 0
			m.updateViewportContent()
		case "backspace":
			// todo
			// m.updateViewportContent()
		case "left":
			// todo
		case "right":
			// todo
		case "up":
			// todo
		case "down":
			// todo
		default:
			// handle regular character input
			// todo
			ch := []rune(msg.String())
			m.buffer.InsertRune(m.cursor.Y, m.cursor.X, ch[0])
			m.cursor.X++
			m.updateViewportContent()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.updateViewportContent()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateViewportContent renders the document content with cursor
func (m *Model) updateViewportContent() {
	var lines []string

	for i, line := range m.buffer.Lines {
		if i == m.cursor.Y {
			// add cursor to current line
			if m.cursor.X <= len(line) {
				cursorLine := string(line[:m.cursor.X]) + "│" + string(line[m.cursor.X:])
				lines = append(lines, cursorLine)
			} else {
				lines = append(lines, string(line)+"│")
			}
		} else {
			lines = append(lines, string(line))
		}
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

func (m Model) headerView() string {
	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("edit")

	divider := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("240")).
		Border(lipgloss.NormalBorder(), false, false, true, false)

	return lipgloss.JoinVertical(lipgloss.Left, header, divider.Render())
}

func (m Model) footerView() string {
	// TODO: maybe use the statusbar?
	totalLines := len(m.buffer.Lines)
	if totalLines == 0 {
		totalLines = 1
	}

	info := fmt.Sprintf("line %d, col %d | lines: %d | press ctrl+c to quit",
		m.cursor.Y+1, m.cursor.X+1, totalLines)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240")).
		Render(info)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  initializing..."
	}
	return lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	)
}
