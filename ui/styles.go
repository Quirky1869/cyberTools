package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Structure définissant la palette de couleurs utilisée pour le thème
type ThemePalette struct {
	Primary    string
	Secondary  string
	Text       string
	Gray       string
	Tertiary   string
	Quaternary string
	Quinary    string
	Senary     string
	Septenary  string
}

// Définition des palettes de couleurs prédéfinies
var (
	NeonTheme = ThemePalette{
		Primary:    "#626262", // Bordures inactives
		Secondary:  "#00f6ff", // Bordure Focus Onglet
		Text:       "#FAFAFA", // Texte principal
		Gray:       "#626262", // Descriptions outils (inactifs)
		Tertiary:   "#FF2A6D", // Onglet Actif (Catégorie sélectionnée)
		Quaternary: "#FF2A6D", // Outil Sélectionné (Flèche et Nom)
		Quinary:    "#FF2A6D", // Description des touches d'aide (Footer)
		Senary:     "#FF2A6D", // Titre ASCII Art
		Septenary:  "#39FF14", // Touches Aide (Footer)
	}

	CyberpunkTheme = ThemePalette{
		Primary:    "#FCEE0A", // Bordures inactives
		Secondary:  "#00F0FF", // Bordure Focus Onglet
		Text:       "#FAFAFA", // Texte principal
		Gray:       "#626262", // Descriptions outils (inactifs)
		Tertiary:   "#FF2A6D", // Onglet Actif
		Quaternary: "#ffa600ff", // Outil Sélectionné
		Quinary:    "#FF2A6D", // Description des touches d'aide
		Senary:     "#00F0FF", // Titre ASCII Art
		Septenary:  "#39FF14", // Touches Aide
	}
)

// Regroupe l'ensemble des styles Lipgloss et Help basés sur la palette active
type Styles struct {
	Doc          lipgloss.Style
	Title        lipgloss.Style
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

// Génère l'ensemble des styles de l'interface à partir de la palette fournie
func MakeStyles(t ThemePalette) Styles {
	return Styles{
		Palette: t,

		Doc: lipgloss.NewStyle().Padding(1, 2),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Senary)).
			Bold(true),

		ActiveTab: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Tertiary)).
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
			Foreground(lipgloss.Color(t.Quaternary)).
			PaddingLeft(1).
			Bold(true).
			SetString("→"),

		Tool: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Text)).
			PaddingLeft(2),

		ToolName: lipgloss.NewStyle().Bold(true),

		ToolDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Gray)),

		Help: help.Styles{
			ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Septenary)),
			FullKey:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Septenary)),
			ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quinary)),
			FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quinary)),
		},
	}
}
