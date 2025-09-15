package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width, height int
	viewport      viewport.Model
	content       []string
	cursorLine    int
	cursorCol     int
	ready         bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// insert new line
			if m.cursorLine < len(m.content) {
				// split current line at cursor position
				currentLine := m.content[m.cursorLine]
				leftPart := currentLine[:m.cursorCol]
				rightPart := currentLine[m.cursorCol:]

				// update current line and insert new line
				m.content[m.cursorLine] = leftPart
				newContent := make([]string, len(m.content)+1)
				copy(newContent[:m.cursorLine+1], m.content[:m.cursorLine+1])
				newContent[m.cursorLine+1] = rightPart
				copy(newContent[m.cursorLine+2:], m.content[m.cursorLine+1:])
				m.content = newContent
			} else {
				m.content = append(m.content, "")
			}
			m.cursorLine++
			m.cursorCol = 0
			m.updateViewportContent()
		case "backspace":
			if m.cursorCol > 0 {
				// remove character before cursor
				line := m.content[m.cursorLine]
				m.content[m.cursorLine] = line[:m.cursorCol-1] + line[m.cursorCol:]
				m.cursorCol--
			} else if m.cursorLine > 0 {
				// join with previous line
				prevLine := m.content[m.cursorLine-1]
				currentLine := m.content[m.cursorLine]
				m.content[m.cursorLine-1] = prevLine + currentLine
				// remove current line
				copy(m.content[m.cursorLine:], m.content[m.cursorLine+1:])
				m.content = m.content[:len(m.content)-1]
				m.cursorLine--
				m.cursorCol = len(prevLine)
			}
			m.updateViewportContent()
		case "left":
			if m.cursorCol > 0 {
				m.cursorCol--
			} else if m.cursorLine > 0 {
				m.cursorLine--
				m.cursorCol = len(m.content[m.cursorLine])
			}
		case "right":
			if m.cursorLine < len(m.content) && m.cursorCol < len(m.content[m.cursorLine]) {
				m.cursorCol++
			} else if m.cursorLine < len(m.content)-1 {
				m.cursorLine++
				m.cursorCol = 0
			}
		case "up":
			if m.cursorLine > 0 {
				m.cursorLine--
				if m.cursorCol > len(m.content[m.cursorLine]) {
					m.cursorCol = len(m.content[m.cursorLine])
				}
			}
		case "down":
			if m.cursorLine < len(m.content)-1 {
				m.cursorLine++
				if m.cursorCol > len(m.content[m.cursorLine]) {
					m.cursorCol = len(m.content[m.cursorLine])
				}
			}
		default:
			// handle regular character input
			if len(msg.String()) == 1 {
				char := msg.String()
				if m.cursorLine >= len(m.content) {
					m.content = append(m.content, "")
				}
				line := m.content[m.cursorLine]
				m.content[m.cursorLine] = line[:m.cursorCol] + char + line[m.cursorCol:]
				m.cursorCol++
				m.updateViewportContent()
			}
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
func (m *model) updateViewportContent() {
	var lines []string

	for i, line := range m.content {
		if i == m.cursorLine {
			// add cursor to current line
			if m.cursorCol <= len(line) {
				cursorLine := line[:m.cursorCol] + "│" + line[m.cursorCol:]
				lines = append(lines, cursorLine)
			} else {
				lines = append(lines, line+"│")
			}
		} else {
			lines = append(lines, line)
		}
	}

	// if cursor is beyond content, add empty lines
	if m.cursorLine >= len(m.content) {
		for i := len(m.content); i <= m.cursorLine; i++ {
			if i == m.cursorLine {
				lines = append(lines, "│")
			} else {
				lines = append(lines, "")
			}
		}
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

func (m model) headerView() string {
	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("edit")

	divider := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("240")).
		// replace --- with something more reasonable here (like a header border)
		Render("─────────────────────────────────────────────────────────────────────────────────")

	return lipgloss.JoinVertical(lipgloss.Left, header, divider)
}

func (m model) footerView() string {
	totalLines := len(m.content)
	if totalLines == 0 {
		totalLines = 1
	}

	info := fmt.Sprintf("line %d, col %d | lines: %d | press ctrl+c to quit",
		m.cursorLine+1, m.cursorCol+1, totalLines)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240")).
		Render(info)
}

func (m model) View() string {
	if !m.ready {
		return "\n  initializing..."
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	)
}

func initialModel() model {
	return model{
		content:    []string{""},
		cursorLine: 0,
		cursorCol:  0,
		ready:      false,
	}
}

func main() {

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
