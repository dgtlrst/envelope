package app

// Text editor using Bubble Tea with optimized cursor rendering

import (
	"envelope/pkg/edit"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
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

	visualCursor cursor.Model
	contentDirty bool
}

func NewModel() Model {
	// Create standard cursor
	visualCursor := cursor.New()
	visualCursor.SetMode(cursor.CursorBlink)
	visualCursor.BlinkSpeed = 500

	visualCursor.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("#00FF00")).
		Foreground(lipgloss.Color("#000000"))
	visualCursor.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	return Model{
		buffer:       edit.NewTextBuffer(),
		cursor:       edit.NewCursor(0, 0),
		visualCursor: visualCursor,
		ready:        false,
		contentDirty: true,
	}
}

func (m *Model) updateViewportContent() {
	// update only if buf content has changed
	if !m.contentDirty {
		return
	}

	var lines = make([]string, len(m.buffer.Lines)) // pre-allocate slice to avoid reallocation
	for i, line := range m.buffer.Lines {
		lines[i] = string(line)
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
	m.contentDirty = false
}

func (m Model) Init() tea.Cmd {
	return m.visualCursor.Focus() // focus the cursor
}

// 1. fix alt and tab presses being considered normal characters
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cursorCmd tea.Cmd

	// update cursor
	m.visualCursor, cursorCmd = m.visualCursor.Update(msg)
	if cursorCmd != nil {
		cmds = append(cmds, cursorCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case (len(msg.Runes) == 1 && msg.Type == tea.KeyRunes) || msg.String() == " ":
			var r rune
			if len(msg.Runes) == 1 {
				r = msg.Runes[0]
			} else {
				r = ' ' // fallback for space
			}
			// m.UndoStack.Push(m.Buffer)
			m.buffer.InsertRune(m.cursor.Y, m.cursor.X, r)
			m.cursor.MoveRight(m.buffer)
			m.contentDirty = true

		case msg.Type == tea.KeyEnter:
			m.buffer.InsertNewLine(m.cursor.Y, m.cursor.X)
			m.cursor.Y++
			m.cursor.X = 0
			m.contentDirty = true

		case msg.Type == tea.KeyBackspace:
			// todo: move this functionality into deleteRune() - this is where these cases should be handled
			if m.cursor.X == 0 && m.cursor.Y > 0 {
				// beginning of line - merge with previous line
				// save length of previous line, if previous line is not empty
				prevLineLen := len(m.buffer.GetLine(m.cursor.Y - 1))
				m.buffer.DeleteRune(m.cursor.Y, m.cursor.X)
				// move cursor to previous line at the junction point
				m.cursor.Y--
				m.cursor.X = prevLineLen
			} else {
				// normal character deletion
				m.buffer.DeleteRune(m.cursor.Y, m.cursor.X)
				m.cursor.X--
			}
			// clamp cursor position to valid range, just in case
			m.cursor.Clamp(m.buffer)
			m.contentDirty = true

		case msg.Type == tea.KeyLeft:
			m.cursor.MoveLeft(m.buffer)

		case msg.Type == tea.KeyRight:
			m.cursor.MoveRight(m.buffer)

		case msg.Type == tea.KeyUp:
			m.cursor.MoveUp(m.buffer)

		case msg.Type == tea.KeyDown:
			m.cursor.MoveDown(m.buffer)

		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit

		default:
			fmt.Println("default case reached:", msg.String())
			log.Println("default case reached on key msg:", msg.String())
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

	// update key press
	m.updateViewportContent()

	// Update viewport
	var viewportCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	if viewportCmd != nil {
		cmds = append(cmds, viewportCmd)
	}

	return m, tea.Batch(cmds...)
}

// set the character under the cursor
func (m *Model) setCursorCharacter() {
	// Always display a thick block cursor
	m.visualCursor.SetChar("█")
}

// render viewport content with cursor overlay
func (m Model) viewportWithCursor() string {
	baseView := m.viewport.View()

	// calculate cursor position in rendered view
	cursorLine := m.cursor.Y - m.viewport.YOffset
	if cursorLine < 0 || cursorLine >= m.viewport.Height {
		return baseView // cursor not visible in viewport
	}

	// split into lines and overlay cursor
	lines := strings.Split(baseView, "\n")
	if cursorLine < len(lines) {
		line := lines[cursorLine]
		runes := []rune(line)

		if m.cursor.X < len(runes) {
			// replace character at cursor position with styled cursor
			before := string(runes[:m.cursor.X])
			after := string(runes[m.cursor.X+1:])
			cursorChar := m.visualCursor.View()
			lines[cursorLine] = before + cursorChar + after
		} else if m.cursor.X == len(runes) {
			// cursor at end of line
			lines[cursorLine] = line + m.visualCursor.View()
		}
	}

	return strings.Join(lines, "\n")
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

	// set character under cursor for visual cursor
	m.setCursorCharacter()

	return lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		m.viewportWithCursor(), // render viewport with cursor overlay
		m.footerView(),
	)
}
