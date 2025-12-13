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
	Septenary  string
}

// 2. Définition des thèmes disponibles
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

// 3. Structure regroupant tous tes styles Lipgloss
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

// 4. Le Générateur : Il fabrique les styles à partir d'une palette donnée
func MakeStyles(t ThemePalette) Styles {
	return Styles{
		Palette: t,

		Doc: lipgloss.NewStyle().Padding(1, 2),

		// Titre ASCII -> Senary
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Senary)).
			Bold(true),

		// 1. Catégories actives -> Tertiary
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

		// 2. Outil Sélectionné -> Quaternary
		SelectedTool: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Quaternary)). // Modification ici
			PaddingLeft(1).
			Bold(true).
			SetString("→"),

		Tool: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Text)).
			PaddingLeft(2),

		ToolName: lipgloss.NewStyle().Bold(true),

		ToolDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Gray)),

		// Styles de l'aide en bas
		Help: help.Styles{
			// Touches -> Septenary
			ShortKey: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Septenary)),
			FullKey:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Septenary)),

			// 3. Descriptions Aide -> Quinary
			ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quinary)), // Modification ici
			FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Quinary)), // Modification ici
		},
	}
}
