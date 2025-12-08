package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Palette Neon Theme
const (
	ColorPrimary   		= "#500aff"
	ColorSecondary 		= "#00f6ff"
	ColorText      		= "#FAFAFA"
	ColorGray      		= "#626262"
	ColorTertiary      	= "#FF2A6D"
	ColorQuaternary     = "#39FF14"
	ColorQuinary    	= "#FFE700"
	ColorSenary    		= "#ff5e00"
)

// Palette Cyberpunk Theme
// const (
// 	ColorPrimary   		= "#FCEE0A" // Cyberpunk Yellow
// 	ColorSecondary 		= "#00F0FF" // Holo Blue
// 	ColorText      		= "#FAFAFA" // White
// 	ColorGray      		= "#626262" // Gray
// 	ColorTertiary      	= "#FF003C" // Glitch Red
// 	ColorQuaternary     = "#39FF14" // Acid Green
// 	ColorQuinary    	= "#FF00FF" // Hot Pink
// 	ColorSenary    		= "#FF8C00" // Sunset Orange
// )

var (
	// Style global pour centrer
	DocStyle = lipgloss.NewStyle().Padding(1, 2)

	// Onglet actif (brillant, couleur secondaire)
	ActiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorSecondary)).
			Padding(0, 3).
			Margin(0, 1).
			MarginTop(1).
			Foreground(lipgloss.Color(ColorTertiary)).
			Bold(true)

	// Onglet inactif (plus sombre, couleur primaire)
	InactiveTabStyle = lipgloss.NewStyle().
		// Border(lipgloss.RoundedBorder()).
		Border(lipgloss.HiddenBorder()).
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
				Foreground(lipgloss.Color(ColorTertiary)).
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
		ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorQuaternary)),
		ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTertiary)),
		FullKey:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorQuaternary)),
		FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTertiary)),
	}
)
