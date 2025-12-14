package main

import (
	"fmt"
	"os"

	"github.com/Quirky1869/cyberTools/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialisation du modèle principal qui contient l'état de l'interface
	m := ui.NewModel()

	// Configuration du programme Bubble Tea avec support de la souris et mode plein écran (AltScreen)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Démarrage de la boucle événementielle de la TUI
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors du lancement de la TUI: %v", err)
		os.Exit(1)
	}
}