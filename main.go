package main

import (
	"envelope/cmd/app"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

func init() {
	logFile, err := os.OpenFile("editor.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("could not open log file:", err)
		os.Exit(1)
	}
	log.SetOutput(logFile)
}

func main() {

	var model = app.NewModel()

	p := tea.NewProgram(&model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	log.Println("Starting program...")

	if _, err := p.Run(); err != nil {
		log.Println("could not start program:", err)
		os.Exit(1)
	}
}
