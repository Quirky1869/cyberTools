package main

import (
	"fmt"
	"os"

	"github.com/Quirky1869/cyberTools/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.NewModel()

	// Activation de la souris et du mode AltScreen (Plein écran)
	p := tea.NewProgram(
		m, 
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Permet de capturer les mouvements souris
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors du lancement de la TUI: %v", err)
		os.Exit(1)
	}
}
