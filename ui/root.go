package ui

import (
	"strings"

	"github.com/Quirky1869/cyberTools/tools"
	aeddsa "github.com/Quirky1869/cyberTools/tools/aed"
	"github.com/Quirky1869/cyberTools/tools/logv"
	"github.com/Quirky1869/cyberTools/tools/sqltui"
	structviewer "github.com/Quirky1869/cyberTools/tools/structViewer"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

// États possibles du focus clavier (Navigation onglets vs Liste outils)
type FocusState int

const (
	FocusCategories FocusState = iota
	FocusTools
)

// Modèle principal contenant l'état global de l'application et l'outil actif
type Model struct {
	categories      []tools.Category
	activeCatIndex  int
	activeToolIndex int
	focus           FocusState
	keys            KeyMap
	help            help.Model
	styles          Styles
	width           int
	height          int
	currentTool     tea.Model
}

// Initialisation du modèle avec chargement du thème et des catégories
func NewModel() Model {
	initialStyles := MakeStyles(NeonTheme)

	h := help.New()
	h.Styles = initialStyles.Help

	return Model{
		categories:  tools.GetCategories(),
		focus:       FocusCategories,
		keys:        DefaultKeyMap,
		help:        h,
		styles:      initialStyles,
		currentTool: nil,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Gestion de l'outil actif : on délègue les événements à l'outil ou on revient au menu (BackMsg)
	if m.currentTool != nil {
		if _, ok := msg.(logv.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		if _, ok := msg.(sqltui.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		if _, ok := msg.(structviewer.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		if _, ok := msg.(aeddsa.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		var cmd tea.Cmd
		m.currentTool, cmd = m.currentTool.Update(msg)
		return m, cmd
	}

	// Gestion du Menu Principal : Navigation et interaction système
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		// Commandes système (Quitter, Aide)
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		// Changement de thème visuel
		case key.Matches(msg, m.keys.ToggleTheme):
			if m.styles.Palette.Primary == NeonTheme.Primary {
				m.styles = MakeStyles(CyberpunkTheme)
			} else {
				m.styles = MakeStyles(NeonTheme)
			}
			m.help.Styles = m.styles.Help

		// Navigation Focus (Onglets <-> Liste)
		case key.Matches(msg, m.keys.Tab):
			if m.focus == FocusCategories {
				m.focus = FocusTools
			} else {
				m.focus = FocusCategories
			}

		// Navigation horizontale (Catégories)
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

		// Navigation verticale (Liste des outils)
		case key.Matches(msg, m.keys.Up):
			if m.focus == FocusTools && m.activeToolIndex > 0 {
				m.activeToolIndex--
			}
		case key.Matches(msg, m.keys.Down):
			if m.focus == FocusTools && m.activeToolIndex < len(m.categories[m.activeCatIndex].Tools)-1 {
				m.activeToolIndex++
			}

		// Initialisation et lancement de l'outil sélectionné
		case key.Matches(msg, m.keys.Enter):
			if m.focus == FocusTools {
				tool := m.categories[m.activeCatIndex].Tools[m.activeToolIndex]

				if tool.Name == "LogV" {
					lv := logv.New(m.width, m.height)
					m.currentTool = lv
					return m, lv.Init()
				}

				if tool.Name == "SqlTUI" {
					st := sqltui.New(m.width, m.height)
					m.currentTool = st
					return m, st.Init()
				}

				if tool.Name == "structViewer" {
					ty := structviewer.New(m.width, m.height)
					m.currentTool = ty
					return m, ty.Init()
				}

				if tool.Name == "AED" {
					a := aeddsa.New(m.width, m.height)
					m.currentTool = a
					return m, a.Init()
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	// Affichage de l'outil actif en plein écran
	if m.currentTool != nil {
		return m.currentTool.View()
	}

	if m.width == 0 {
		return "loading..."
	}

	// Génération du titre ASCII art
	titleRaw := figure.NewFigure("cyberTools", "slant", true).String()
	titleArt := m.styles.Title.Render(titleRaw)

	centeredTitle := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		titleArt,
	)

	// Construction de la barre d'onglets (Catégories)
	var tabs []string
	for i, cat := range m.categories {
		style := m.styles.InactiveTab

		if m.activeCatIndex == i {
			style = m.styles.ActiveTab
			if m.focus == FocusCategories {
				style = style.BorderForeground(lipgloss.Color(m.styles.Palette.Secondary))
			} else {
				style = style.BorderForeground(lipgloss.Color(m.styles.Palette.Gray))
			}
		}
		tabs = append(tabs, style.Render(cat.Name))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Construction de la liste des outils
	currentTools := m.categories[m.activeCatIndex].Tools
	var toolList strings.Builder

	for i, tool := range currentTools {
		style := m.styles.Tool
		cursor := "  "

		if m.activeToolIndex == i {
			style = m.styles.SelectedTool
		}

		name := m.styles.ToolName.Render(tool.Name)
		desc := m.styles.ToolDesc.Render("(" + tool.Description + ")")
		toolList.WriteString(style.Render(cursor+name+" "+desc) + "\n")
	}

	// Configuration du conteneur principal avec bordures dynamiques selon le focus
	const menuHeight = 15
	boxBorderColor := lipgloss.Color(m.styles.Palette.Primary)
	if m.focus == FocusTools {
		boxBorderColor = lipgloss.Color(m.styles.Palette.Secondary)
	}

	contentBox := m.styles.Window.
		Width(max(lipgloss.Width(row), 60)).
		Height(menuHeight).
		BorderForeground(boxBorderColor).
		Render(toolList.String())

	// Assemblage final de l'interface graphique
	helpView := m.help.View(m.keys)
	appContent := lipgloss.JoinVertical(lipgloss.Left, row, contentBox, "\n"+helpView)

	centeredAppContent := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		appContent,
	)

	finalUI := lipgloss.JoinVertical(
		lipgloss.Left,
		centeredTitle,
		"\n",
		centeredAppContent,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		finalUI,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
