package ui

import (
    "strings"

    "github.com/Quirky1869/cyberTools/tools"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    figure "github.com/common-nighthawk/go-figure"
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
    // Champs pour l'aide
    keys            KeyMap
    help            help.Model
    width           int
    height          int
}

func NewModel() Model {
    helpPaint := help.New()
    helpPaint.Styles = HelpStyles
    return Model{
        categories: tools.GetCategories(),
        focus:      FocusCategories,
        keys:       DefaultKeyMap,
        help:       helpPaint,
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
        m.help.Width = msg.Width

    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Help):
            m.help.ShowAll = !m.help.ShowAll
        case key.Matches(msg, m.keys.Tab):
            if m.focus == FocusCategories {
                m.focus = FocusTools
            } else {
                m.focus = FocusCategories
            }
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

// --------------------------------------------------------------------------------
// VIEW FUNCTION (RENDU VISUEL)
// --------------------------------------------------------------------------------

func (m Model) View() string {
    if m.width == 0 {
        return "loading..."
    }

    // --- 0. Génération et centrage du titre ---
    titleArt := generateTitle()
    centeredTitle := lipgloss.PlaceHorizontal(
        m.width, // Largeur totale de l'écran
        lipgloss.Center,
        titleArt,
    )

    // 1. Navbar (Onglets)
    var tabs []string
    for i, cat := range m.categories {
        style := InactiveTabStyle

        if m.activeCatIndex == i {
            style = ActiveTabStyle.Copy()
            if m.focus == FocusCategories {
                // Bordure couleur cyan de l'onglet actif
                style = style.BorderForeground(lipgloss.Color(ColorSecondary))
            } else {
                // Bordure couleur gris quand on fait "tab" pour aller dans les outils
                style = style.BorderForeground(lipgloss.Color(ColorGray))
            }
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
        }
        name := ToolNameStyle.Render(tool.Name)
        desc := ToolDescStyle.Render("(" + tool.Description + ")")
        toolList.WriteString(style.Render(cursor + name + " " + desc) + "\n")
    }

    const menuHeight = 15

    contentBox := WindowStyle.
        Width(max(lipgloss.Width(row), 60)).
        Height(menuHeight).
        Render(toolList.String())

    if m.focus == FocusTools {
        contentBox = WindowStyle.
            BorderForeground(lipgloss.Color(ColorSecondary)).
            Width(max(lipgloss.Width(row), 60)).
            Height(menuHeight).
            Render(toolList.String())
    }

    // 3. Section Aide
    helpView := m.help.View(m.keys)

    // 4. Assemblage du contenu principal (Navbar + Box + Aide)
    appContent := lipgloss.JoinVertical(lipgloss.Left, row, contentBox, "\n"+helpView)

    // On centre l'application (navbar, contenu, aide)
    centeredAppContent := lipgloss.PlaceHorizontal(
        m.width,
        lipgloss.Center,
        appContent,
    )

    // 5. Assemblage final : Titre + Contenu Centré
    finalUI := lipgloss.JoinVertical(
        lipgloss.Left,
        centeredTitle,
        "\n", // Espace entre le titre et les onglets
        centeredAppContent,
    )

    // On centre le bloc final verticalement
    return lipgloss.Place(
        m.width,
        m.height,
        lipgloss.Center, // Centrage horizontal
        lipgloss.Center, // Centrage vertical
        finalUI,
    )
}

//   ____________________  ______              __             ______            __    
//  /_  __/  _/_  __/ __ \/ ____/  _______  __/ /_  ___  ____/_  __/___  ____  / /____
//   / /  / /  / / / /_/ / __/    / ___/ / / / __ \/ _ \/ ___// / / __ \/ __ \/ / ___/
//  / / _/ /  / / / _, _/ /___   / /__/ /_/ / /_/ /  __/ /   / / / /_/ / /_/ / (__  ) 
// /_/ /___/ /_/ /_/ |_/_____/   \___/\__, /_.___/\___/_/   /_/  \____/\____/_/____/  
//                                   /____/                                           

// generateTitle crée le titre "cyberTools" en ASCII Art stylisé
func generateTitle() string {
    figure := figure.NewFigure("cyberTools", "slant", true)
    title := figure.String()

	styledTitle := lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColorTertiary)).
        // Background(lipgloss.Color(ColorPrimary)).
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
