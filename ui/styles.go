package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// 1. Définition de la structure d'une Palette de couleurs
type ThemePalette struct {
	Primary    string
	Secondary  string
	Text       string
	Gray       string
	Tertiary   string
	Quaternary string
	Quinary    string
	Senary     string
}

// 2. Définition des thèmes disponibles
var (
	NeonTheme = ThemePalette{
		Primary:    "#500aff",
		Secondary:  "#00f6ff",
		Text:       "#FAFAFA",
		Gray:       "#626262",
		Tertiary:   "#FF2A6D",
		Quaternary: "#39FF14",
		Quinary:    "#FFE700",
		Senary:     "#ff5e00",
	}

	CyberpunkTheme = ThemePalette{
		Primary:    "#FCEE0A", // Jaune
		Secondary:  "#00F0FF", // Bleu Holo
		Text:       "#FAFAFA",
		Gray:       "#626262",
		Tertiary:   "#FF003C", // Rouge Glitch
		Quaternary: "#39FF14", // Vert Acide
		Quinary:    "#FF00FF", // Rose Hot
		Senary:     "#FF8C00", // Orange
	}
)

// 3. Structure regroupant tous tes styles Lipgloss
// Au lieu de variables globales, on les met dans une boîte
type Styles struct {
	Doc          lipgloss.Style
	ActiveTab    lipgloss.Style
	InactiveTab  lipgloss.Style
	Window       lipgloss.Style
	SelectedTool lipgloss.Style
	Tool         lipgloss.Style
	ToolName     lipgloss.Style
	ToolDesc     lipgloss.Style
	Help         help.Styles
	Palette      ThemePalette
}

// 4. Le Générateur : Il fabrique les styles à partir d'une palette donnée
func MakeStyles(t ThemePalette) Styles {
	return Styles{
		Palette: t,

		Doc: lipgloss.NewStyle().Padding(1, 2),

		ActiveTab: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Secondary)).
			Padding(0, 3).
			Margin(0, 1).
			MarginTop(1).
			Foreground(lipgloss.Color(t.Tertiary)).
			Bold(true),

		InactiveTab: lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder()).
			BorderForeground(lipgloss.Color(t.Primary)).
			Padding(0, 1).
			Foreground(lipgloss.Color(t.Gray)),

		Window: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Primary)).
			Padding(1, 2).
			Align(lipgloss.Left),

		SelectedTool: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Tertiary)).
			PaddingLeft(1).
			Bold(true).
			SetString("→"),

		Tool: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Text)).
			PaddingLeft(2),

		ToolName: lipgloss.NewStyle().Bold(true),

		ToolDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Gray)),

		Help: help.Styles{
			ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quaternary)),
			ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Tertiary)),
			FullKey:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quaternary)),
			FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Tertiary)),
		},
	}
}
