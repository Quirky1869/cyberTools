package ui

import (
	"strings"

	"github.com/Quirky1869/cyberTools/tools"
	"github.com/Quirky1869/cyberTools/tools/logv"
	"github.com/Quirky1869/cyberTools/tools/sqltui"
	"github.com/Quirky1869/cyberTools/tools/structViewer"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/Quirky1869/cyberTools/tools/aed"
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
	keys            KeyMap
	help            help.Model
	styles          Styles
	width           int
	height          int

	// Champ pour stocker l'outil en cours d'exécution (nil si on est au menu)
	currentTool tea.Model
}

func NewModel() Model {
	// On initialise les styles avec le thème Neon par défaut
	initialStyles := MakeStyles(NeonTheme)

	// On configure l'aide avec les styles du thème
	h := help.New()
	h.Styles = initialStyles.Help

	return Model{
		categories:  tools.GetCategories(),
		focus:       FocusCategories,
		keys:        DefaultKeyMap,
		help:        h,
		styles:      initialStyles,
		currentTool: nil, // Pas d'outil au démarrage
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ---------------------------------------------------------
	// 1. GESTION DE L'OUTIL ACTIF (SI UN OUTIL EST LANCÉ)
	// ---------------------------------------------------------
	if m.currentTool != nil {
		// On vérifie si l'outil envoie le signal de retour "BackMsg"
		if _, ok := msg.(logv.BackMsg); ok {
			m.currentTool = nil       // On ferme l'outil
			return m, tea.ClearScreen // On nettoie l'écran pour réafficher le menu proprement
		}

		if _, ok := msg.(sqltui.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		if _, ok := msg.(structviewer.BackMsg); ok {
			m.currentTool = nil
			return m, tea.ClearScreen
		}

		if _, ok := msg.(aed.BackMsg); ok {
    	m.currentTool = nil
    	return m, tea.ClearScreen
}

		// Sinon, on transmet le message à l'outil pour qu'il le gère
		var cmd tea.Cmd
		m.currentTool, cmd = m.currentTool.Update(msg)
		return m, cmd
	}

	// ---------------------------------------------------------
	// 2. GESTION DU MENU PRINCIPAL (SI AUCUN OUTIL N'EST LANCÉ)
	// ---------------------------------------------------------
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		// --- Touches Système ---
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		// --- Changement de Thème (Touche 't') ---
		case key.Matches(msg, m.keys.ToggleTheme):
			// Bascule simple entre Neon et Cyberpunk
			if m.styles.Palette.Primary == NeonTheme.Primary {
				m.styles = MakeStyles(CyberpunkTheme)
			} else {
				m.styles = MakeStyles(NeonTheme)
			}
			// Important : On met à jour les styles de l'aide immédiatement
			m.help.Styles = m.styles.Help

		// --- Navigation Focus (Tab) ---
		case key.Matches(msg, m.keys.Tab):
			if m.focus == FocusCategories {
				m.focus = FocusTools
			} else {
				m.focus = FocusCategories
			}

		// --- Navigation Gauche/Droite (Catégories) ---
		case key.Matches(msg, m.keys.Left):
			if m.focus == FocusCategories && m.activeCatIndex > 0 {
				m.activeCatIndex--
				m.activeToolIndex = 0 // Reset de la sélection outil
			}
		case key.Matches(msg, m.keys.Right):
			if m.focus == FocusCategories && m.activeCatIndex < len(m.categories)-1 {
				m.activeCatIndex++
				m.activeToolIndex = 0
			}

		// --- Navigation Haut/Bas (Outils) ---
		case key.Matches(msg, m.keys.Up):
			if m.focus == FocusTools && m.activeToolIndex > 0 {
				m.activeToolIndex--
			}
		case key.Matches(msg, m.keys.Down):
			if m.focus == FocusTools && m.activeToolIndex < len(m.categories[m.activeCatIndex].Tools)-1 {
				m.activeToolIndex++
			}

		// --- Lancement d'un outil (Entrée) ---
		case key.Matches(msg, m.keys.Enter):
			if m.focus == FocusTools {
				// On récupère l'outil actuellement sélectionné
				tool := m.categories[m.activeCatIndex].Tools[m.activeToolIndex]

				// Si c'est "LogV", on initialise son modèle
				if tool.Name == "LogV" {
					// IMPORTANT : On passe la taille de la fenêtre actuelle (m.width, m.height)
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
    				a := aed.New(m.width, m.height)
    				m.currentTool = a
    				return m, a.Init()
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	// A. Si un outil est actif, on affiche SA vue à lui (plein écran)
	if m.currentTool != nil {
		return m.currentTool.View()
	}

	// B. Sinon, on affiche le MENU PRINCIPAL
	if m.width == 0 {
		return "loading..."
	}

	// --- 0. Génération et centrage du titre ---
	titleArt := generateTitle(m.styles.Palette)
	centeredTitle := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		titleArt,
	)

	// --- 1. Navbar (Onglets) ---
	var tabs []string
	for i, cat := range m.categories {
		// Par défaut : Inactif
		style := m.styles.InactiveTab

		// Si c'est l'onglet courant
		if m.activeCatIndex == i {
			// On prend le style Actif
			style = m.styles.ActiveTab

			// Gestion fine de la bordure selon le Focus
			if m.focus == FocusCategories {
				// Focus sur les onglets -> Bordure Secondaire (Cyan)
				style = style.BorderForeground(lipgloss.Color(m.styles.Palette.Secondary))
			} else {
				// Focus ailleurs -> Bordure Grise
				style = style.BorderForeground(lipgloss.Color(m.styles.Palette.Gray))
			}
		}
		tabs = append(tabs, style.Render(cat.Name))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// --- 2. Contenu (Liste outils) ---
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

	const menuHeight = 15

	// Couleur de la bordure de la boîte
	boxBorderColor := lipgloss.Color(m.styles.Palette.Primary) // Par défaut (Violet/Jaune)
	if m.focus == FocusTools {
		boxBorderColor = lipgloss.Color(m.styles.Palette.Secondary) // Si Focus (Cyan/Bleu)
	}

	contentBox := m.styles.Window.
		Width(max(lipgloss.Width(row), 60)).
		Height(menuHeight).
		BorderForeground(boxBorderColor).
		Render(toolList.String())

	// --- 3. Section Aide ---
	helpView := m.help.View(m.keys)

	// --- 4. Assemblage Final ---
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

// generateTitle crée le titre en ASCII Art avec la palette fournie
func generateTitle(p ThemePalette) string {
	figure := figure.NewFigure("cyberTools", "slant", true)
	title := figure.String()

	styledTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Tertiary)).
		Bold(true).
		Render(title)

	return styledTitle
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
