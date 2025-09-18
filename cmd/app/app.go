package app

// todo:
// 1. updateViewportContent() is called too often, needs to be optimized
//    also it currently is the one way to update "visual" positions

import (
	"envelope/pkg/edit"
	"fmt"
	"log"
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

	baseContent   string   // content without cursosr
	contentDirty  bool     // track if buffer content changed
	lastCursorPos struct { // track cursor position changes
		X, Y int
	}
	renderedContent string // final content with cursor (cached)
}

func NewModel() Model {
	return Model{
		buffer:       edit.NewTextBuffer(),
		cursor:       edit.NewCursor(0, 0),
		ready:        false,
		contentDirty: true, // force initial render
		lastCursorPos: struct {
			X int
			Y int
		}{-1, -1}, // invalid initial position
	}
}

func (m *Model) generateBaseContent() {
	if !m.contentDirty {
		return // skip if content hasn't changed
	}

	var lines []string
	for _, line := range m.buffer.Lines {
		lines = append(lines, string(line))
	}
	m.baseContent = strings.Join(lines, "\n")
	m.contentDirty = false
}

// adds cursor to base content at current position
func (m *Model) injectCursor() string {
	if m.baseContent == "" {
		return ""
	}

	// Safety check for cursor position
	if m.cursor.Y >= len(m.buffer.Lines) {
		return m.baseContent
	}

	// Find the start position of the cursor line in baseContent
	lineStart := 0
	for i := 0; i < m.cursor.Y; i++ {
		lineStart += len(string(m.buffer.Lines[i])) + 1 // +1 for newline
	}

	// Find the end position of the cursor line
	currentLineStr := string(m.buffer.Lines[m.cursor.Y])
	lineEnd := lineStart + len(currentLineStr)

	// Calculate cursor position within the line
	cursorPos := lineStart + m.cursor.X
	if cursorPos > lineEnd {
		cursorPos = lineEnd
	}

	// Build result by inserting cursor at the right position
	var result strings.Builder
	result.Grow(len(m.baseContent) + 3) // Pre-allocate space

	// Add content before cursor
	result.WriteString(m.baseContent[:cursorPos])

	// Add cursor
	result.WriteString("│")

	// Add content after cursor
	result.WriteString(m.baseContent[cursorPos:])

	return result.String()
}

func (m *Model) updateViewportContentSmart() {
	// check if cursor position changed
	cursorMoved := m.cursor.X != m.lastCursorPos.X || m.cursor.Y != m.lastCursorPos.Y

	// only update if content or cursor changed
	if !m.contentDirty && !cursorMoved {
		return
	}

	// Update base content if buffer changed
	m.generateBaseContent()

	// Generate final content with cursor
	finalContent := m.injectCursor()

	// Only call SetContent if content actually changed
	if finalContent != m.renderedContent {
		m.viewport.SetContent(finalContent)
		m.renderedContent = finalContent
	}

	// Update cursor position tracking
	m.lastCursorPos.X = m.cursor.X
	m.lastCursorPos.Y = m.cursor.Y
}

// markContentDirty flags that buffer content has changed
func (m *Model) markContentDirty() {
	m.contentDirty = true
}

func (m Model) Init() tea.Cmd {
	return nil
}

// 1. fix alt and tab presses being considered normal characters
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
			m.markContentDirty() // NEW LINE

		case msg.Type == tea.KeyEnter:
			m.buffer.InsertNewLine(m.cursor.Y, m.cursor.X)
			m.cursor.Y++
			m.cursor.X = 0
			m.markContentDirty()

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
			m.markContentDirty()

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
			m.updateViewportContentSmart()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// Only update viewport content when necessary
	switch msg.(type) {
	case tea.KeyMsg, tea.WindowSizeMsg:
		m.updateViewportContentSmart()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
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
