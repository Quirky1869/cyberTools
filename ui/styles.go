package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Palette
const (
	ColorPrimary   = "#500aff"
	ColorSecondary = "#00f6ff"
	ColorText      = "#FAFAFA"
	ColorGray      = "#626262"
	ColorPink      = "#FF2A6D"
	ColorGreen     = "#39FF14"
	ColorYellow    = "#FFE700"
	ColorOrange    = "#ff5e00"
)

var (
	// Style global pour centrer
	DocStyle = lipgloss.NewStyle().Padding(1, 2)

	// Onglet actif (brillant, couleur secondaire)
	ActiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorSecondary)).
			Padding(0, 1).
			Foreground(lipgloss.Color(ColorPink)).
			Bold(true)

	// Onglet inactif (plus sombre, couleur primaire)
	InactiveTabStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorPrimary)).
				Padding(0, 1).
				Foreground(lipgloss.Color(ColorGray))

	// --- Style du contenu (Liste des outils) ---

	// La fenêtre principale qui contient la liste
	WindowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPrimary)).
			Padding(1, 2).
			Align(lipgloss.Left)

	// Outil sélectionné
	SelectedToolStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorPink)).
				PaddingLeft(1).
				Bold(true).
				SetString("→") // Ajoute une flèche devant

	// Outil non sélectionné
	ToolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			PaddingLeft(2) // Pour aligner avec le curseur du dessus

	// Titre de l'outil (Nom)
	ToolNameStyle = lipgloss.NewStyle().Bold(true)

	// Description des outils (plus discret)
	ToolDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray))

	// Défini les couleurs de l'aide (en bas de la TUI)
	HelpStyles = help.Styles{
		ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)),
		ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPink)),
		FullKey:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)),
		FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPink)),
	}
)
