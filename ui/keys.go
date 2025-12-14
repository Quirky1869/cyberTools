package ui

import "github.com/charmbracelet/bubbles/key"

// Définition de tous les raccourcis clavier disponibles dans l'application
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Tab         key.Binding
	Enter       key.Binding
	Help        key.Binding
	Quit        key.Binding
	ToggleTheme key.Binding
}

// Implémentation de l'interface d'aide
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.ToggleTheme, k.Tab, k.Enter},
		{k.Help, k.Quit},
	}
}

// Configuration des touches par défaut (support navigation Vim et flèches)
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "haut"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "bas"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "catégorie préc."),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "catégorie suiv."),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "changer section"),
	),
	ToggleTheme: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "changer thème"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "valider"),
	),
	Help: key.NewBinding(
		key.WithKeys("?", ","),
		key.WithHelp("?", "aide"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quitter"),
	),
}
