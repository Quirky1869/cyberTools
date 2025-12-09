package logv

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ncruces/zenity"
)

// BackMsg est un message spécial pour dire à root.go : "C'est fini, reviens au menu"
type BackMsg struct{}

// Les états internes de l'outil
type SessionState int

const (
	StateChooseMethod SessionState = iota
	StatePickingFile
	StateViewing
)

// Styles locaux (tu peux importer ui si tu veux, mais c'est bien de garder l'outil autonome)
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Pink
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true) // Rouge
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true) // Orange
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))            // Cyan
)

type Model struct {
	state        SessionState
	filePicker   filepicker.Model
	viewport     viewport.Model
	textInput    textinput.Model
	selectedFile string
	content      []string // Lignes brutes
	filtered     []int    // Index des lignes affichées
	filtering    bool
	width        int
	height       int
}

func New() Model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".log", ".txt", ".go", ".md"}
	fp.CurrentDirectory, _ = os.Getwd()

	ti := textinput.New()
	ti.Placeholder = "Filtrer (ex: /db)..."
	ti.CharLimit = 156
	ti.Width = 30

	return Model{
		state:      StateChooseMethod,
		filePicker: fp,
		textInput:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4 // Espace pour header/footer
		m.filePicker.Height = msg.Height - 5

	case tea.KeyMsg:
		// Gestion globale de la sortie (si on n'est pas en train d'écrire)
		if !m.filtering && (msg.String() == "q" || msg.String() == "ctrl+c") {
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	// Machine à état
	switch m.state {

	case StateChooseMethod:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "t": // Terminal
				m.state = StatePickingFile
				m.filePicker.CurrentDirectory, _ = os.Getwd()
				return m, m.filePicker.Init()
			case "g": // Graphique
				// Note: Zenity bloque un peu l'interface le temps de la sélection
				path, err := zenity.SelectFile(zenity.Filename(m.filePicker.CurrentDirectory + "/"))
				if err == nil && path != "" {
					return m.loadFile(path)
				}
			}
		}

	case StatePickingFile:
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			return m.loadFile(path)
		}

	case StateViewing:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if m.filtering {
				switch msg.String() {
				case "enter", "esc":
					m.filtering = false
					m.textInput.Blur()
					m.applyFilter() // Appliquer le filtre saisi
				}
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

			switch msg.String() {
			case "/":
				m.filtering = true
				m.textInput.Focus()
				return m, textinput.Blink
			case "backspace":
				m.textInput.Reset()
				m.applyFilter()
			case "n", "p":
				// Implémentation simple : scroll page par page pour l'instant
				// (La navigation précise next/prev demande une logique d'index plus complexe)
			case "esc":
				// Retour au choix fichier si on veut, ou quitter
				return m, func() tea.Msg { return BackMsg{} }
			}
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case StateChooseMethod:
		return fmt.Sprintf(
			"\n  %s\n\n  [T] Choisir depuis le Terminal\n  [G] Choisir depuis une vue Graphique\n\n  [Q] Retour Menu",
			titleStyle.Render("LogV - Importation"),
		)

	case StatePickingFile:
		return "\n" + m.filePicker.View()

	case StateViewing:
		header := titleStyle.Render("LogV: " + m.selectedFile)
		footer := infoStyle.Render("\n[ / ] Filtrer  [ Bksp ] Reset Filtre  [ q ] Quitter")

		if m.filtering {
			footer = fmt.Sprintf("\nFiltre : %s", m.textInput.View())
		}

		return fmt.Sprintf("%s\n%s%s", header, m.viewport.View(), footer)
	}
	return ""
}

// loadFile lit le fichier et prépare le viewport
func (m Model) loadFile(path string) (Model, tea.Cmd) {
	m.selectedFile = path
	content, err := os.ReadFile(path)
	if err != nil {
		return m, nil // Gérer l'erreur idéalement
	}
	m.content = strings.Split(string(content), "\n")
	m.viewport = viewport.New(m.width, m.height-4)
	m.applyFilter()
	m.state = StateViewing
	return m, nil
}

// applyFilter génère le contenu coloré et filtré
func (m *Model) applyFilter() {
	var builder strings.Builder
	filter := strings.ToLower(m.textInput.Value())

	for _, line := range m.content {
		// 1. Filtrage
		if filter != "" && !strings.Contains(strings.ToLower(line), filter) {
			continue
		}

		// 2. Coloriage
		coloredLine := line
		if strings.Contains(line, "ERROR") {
			coloredLine = errorStyle.Render(line)
		} else if strings.Contains(line, "WARN") {
			coloredLine = warnStyle.Render(line)
		}

		builder.WriteString(coloredLine + "\n")
	}
	m.viewport.SetContent(builder.String())
	m.viewport.GotoTop()
}