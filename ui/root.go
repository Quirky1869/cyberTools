package ui

import (
	"strings"

	"github.com/Quirky1869/cyberTools/tools"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusState int

const (
	FocusCategories FocusState = iota
	FocusTools
)

type Model struct {
	categories      []tools.Category
	activeCatIndex  int
	activeToolIndex int
	
	focus           FocusState
	
	// Nouveaux champs pour l'aide
	keys            KeyMap
	help            help.Model
	
	width           int
	height          int
}

func NewModel() Model {
	return Model{
		categories: tools.GetCategories(),
		focus:      FocusCategories,
		// Initialisation des touches et de l'aide
		keys:       DefaultKeyMap,
		help:       help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width // Adapter la largeur de l'aide

	case tea.KeyMsg:
		switch {
		// --- Touches globales ---
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll // Bascule entre aide courte/longue

		case key.Matches(msg, m.keys.Tab):
			if m.focus == FocusCategories {
				m.focus = FocusTools
			} else {
				m.focus = FocusCategories
			}

		// --- Navigation ---
		case key.Matches(msg, m.keys.Left):
			if m.focus == FocusCategories && m.activeCatIndex > 0 {
				m.activeCatIndex--
				m.activeToolIndex = 0
			}

		case key.Matches(msg, m.keys.Right):
			if m.focus == FocusCategories && m.activeCatIndex < len(m.categories)-1 {
				m.activeCatIndex++
				m.activeToolIndex = 0
			}

		case key.Matches(msg, m.keys.Up):
			if m.focus == FocusTools && m.activeToolIndex > 0 {
				m.activeToolIndex--
			}

		case key.Matches(msg, m.keys.Down):
			if m.focus == FocusTools && m.activeToolIndex < len(m.categories[m.activeCatIndex].Tools)-1 {
				m.activeToolIndex++
			}

		case key.Matches(msg, m.keys.Enter):
			if m.focus == FocusTools {
				// Action à venir
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	// 1. Navbar (Onglets)
	var tabs []string
	for i, cat := range m.categories {
		style := InactiveTabStyle
		if m.activeCatIndex == i {
			style = ActiveTabStyle
		}
		tabs = append(tabs, style.Render(cat.Name))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// 2. Contenu (Liste outils)
	currentTools := m.categories[m.activeCatIndex].Tools
	var toolList strings.Builder

	for i, tool := range currentTools {
		style := ToolStyle
		cursor := "  "
		if m.activeToolIndex == i {
			style = SelectedToolStyle
			cursor = "> "
		}
		name := ToolNameStyle.Render(tool.Name)
		desc := ToolDescStyle.Render("(" + tool.Description + ")")
		toolList.WriteString(style.Render(cursor + name + " " + desc) + "\n")
	}

	// Calcul de la hauteur disponible pour le contenu
	// Hauteur totale - Hauteur Onglets - Hauteur Aide - Padding
	const menuHeight = 15 

    // 2. On applique cette hauteur à la boîte
    contentBox := WindowStyle.
        Width(max(lipgloss.Width(row), 60)).
        Height(menuHeight). // <-- Hauteur fixe ici !
        Render(toolList.String())

    if m.focus == FocusTools {
        contentBox = WindowStyle.
            BorderForeground(lipgloss.Color(ColorSecondary)).
            Width(max(lipgloss.Width(row), 60)).
            Height(menuHeight). // <-- Ici aussi
            Render(toolList.String())
    }

	// 3. Section Aide
	helpView := m.help.View(m.keys)
	
	// 4. Assemblage final
	// On empile : Navbar -> Contenu -> Aide
	ui := lipgloss.JoinVertical(lipgloss.Left, row, contentBox, "\n"+helpView)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		ui,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}