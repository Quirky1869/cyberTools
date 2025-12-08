package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap définit toutes les touches utilisables
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding
	Enter key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp renvoie les touches affichées tout le temps en bas
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Quit}
}

// FullHelp renvoie toutes les touches (affiché quand on appuie sur ?)
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // Navigation
		{k.Tab, k.Enter},                // Actions
		{k.Help, k.Quit},                // Système
	}
}

// DefaultKeyMap définit les touches par défaut
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
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "valider"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "aide"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quitter"),
	),
}
