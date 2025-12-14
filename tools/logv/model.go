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

// Message de retour pour signaler au menu principal de reprendre la main
type BackMsg struct{}

// États de la machine à états interne
type SessionState int

const (
	StateChooseMethod SessionState = iota
	StatePickingFile
	StateViewing
)

// Définition des styles de l'interface
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))
	pathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#500aff")).Bold(true)
)

// Modèle principal contenant les composants (Picker, Viewport) et l'état des données
type Model struct {
	state        SessionState
	filePicker   filepicker.Model
	viewport     viewport.Model
	textInput    textinput.Model
	pathInput    textinput.Model
	selectedFile string
	content      []string
	width        int
	height       int
	filtering    bool
	enteringPath bool
}

// Initialisation des composants avec configuration des couleurs et dimensions
func New(w, h int) Model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".log", ".txt", ".go", ".md", ".json", ".yaml", ".conf"}
	fp.CurrentDirectory, _ = os.Getwd()
	fp.Height = h - 8
	fp.ShowHidden = false

	// Personnalisation des couleurs du FilePicker
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))
	fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Input pour le filtre de contenu
	tiFilter := textinput.New()
	tiFilter.Placeholder = "Filtrer (ex: /db)..."
	tiFilter.CharLimit = 156
	tiFilter.Width = 30

	// Input pour la saisie manuelle de chemin
	tiPath := textinput.New()
	tiPath.Placeholder = "/home/user/..."
	tiPath.CharLimit = 256
	tiPath.Width = 50

	vp := viewport.New(w, h-4)

	return Model{
		state:        StateChooseMethod,
		filePicker:   fp,
		textInput:    tiFilter,
		pathInput:    tiPath,
		viewport:     vp,
		width:        w,
		height:       h,
		enteringPath: false,
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
		m.viewport.Height = msg.Height - 4
		m.filePicker.Height = msg.Height - 8

	case tea.KeyMsg:
		// Sortie globale si aucune saisie n'est en cours
		if !m.filtering && !m.enteringPath && (msg.String() == "ctrl+c") {
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	switch m.state {

	// Écran de choix : Terminal vs GUI
	case StateChooseMethod:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q":
				return m, func() tea.Msg { return BackMsg{} }
			case "t":
				m.state = StatePickingFile
				m.filePicker.CurrentDirectory, _ = os.Getwd()
				return m, m.filePicker.Init()
			case "g":
				path, err := zenity.SelectFile(zenity.Filename(m.filePicker.CurrentDirectory + "/"))
				if err == nil && path != "" {
					return m.loadFile(path)
				}
			}
		}

	// Écran de sélection de fichier (File Picker)
	case StatePickingFile:
		// Gestion de la saisie manuelle du chemin (Ctrl+L)
		if m.enteringPath {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "enter":
					newPath := m.pathInput.Value()
					info, err := os.Stat(newPath)
					if err == nil && info.IsDir() {
						m.filePicker.CurrentDirectory = newPath
						m.filePicker.Init()
					}
					m.enteringPath = false
					m.pathInput.Blur()
				case "esc":
					m.enteringPath = false
					m.pathInput.Blur()
				}
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

		// Navigation standard dans le File Picker
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q":
				m.state = StateChooseMethod
				return m, nil
			case "h":
				m.filePicker.ShowHidden = !m.filePicker.ShowHidden
				return m, m.filePicker.Init()
			case "ctrl+l":
				m.enteringPath = true
				m.pathInput.SetValue(m.filePicker.CurrentDirectory)
				m.pathInput.Focus()
				return m, textinput.Blink
			}
		}

		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			return m.loadFile(path)
		}

	// Écran de visualisation du fichier (Viewer)
	case StateViewing:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Gestion de la barre de filtrage
			if m.filtering {
				switch msg.String() {
				case "enter", "esc":
					m.filtering = false
					m.textInput.Blur()
					m.applyFilter()
				}
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

			// Commandes du viewer
			switch msg.String() {
			case "/":
				m.filtering = true
				m.textInput.Focus()
				return m, textinput.Blink
			case "backspace":
				m.textInput.Reset()
				m.applyFilter()
			case "esc", "q":
				m.state = StatePickingFile
				return m, nil
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
			"\n  %s\n\n  [t] Choisir depuis le Terminal\n  [g] Choisir depuis une vue Graphique\n\n  [q] Retour Menu",
			titleStyle.Render("LogV - Importation"),
		)

	case StatePickingFile:
		title := titleStyle.Render("LogV - Ouvrir un fichier")
		currentDir := fmt.Sprintf(" %s", pathStyle.Render(m.filePicker.CurrentDirectory))

		// Affichage conditionnel selon si on tape un chemin ou si on navigue
		var content string
		if m.enteringPath {
			content = fmt.Sprintf("\n\nEntrez le chemin absolu :\n%s", m.pathInput.View())
		} else {
			content = "\n" + m.filePicker.View()
		}

		var helpText string
		if m.enteringPath {
			helpText = "enter: valider • esc: annuler"
		} else {
			hiddenStatus := "off"
			if m.filePicker.ShowHidden {
				hiddenStatus = "ON"
			}
			helpText = fmt.Sprintf("↑/↓/←/→: naviguer • enter: ouvrir • h: cachés(%s) • ctrl+l: chemin • esc: retour", hiddenStatus)
		}

		footer := helpStyle.Render("\n" + helpText)

		return fmt.Sprintf("\n  %s\n\n  %s%s%s", title, currentDir, content, footer)

	case StateViewing:
		header := titleStyle.Render("LogV: " + m.selectedFile)
		footer := infoStyle.Render("\n[ / ] Filtrer  [ Bksp ] Reset Filtre  [ q ] Retour")

		if m.filtering {
			footer = fmt.Sprintf("\nFiltre : %s", m.textInput.View())
		}

		return fmt.Sprintf("%s\n%s%s", header, m.viewport.View(), footer)
	}
	return ""
}

// Charge le fichier en mémoire et applique le filtre initial
func (m Model) loadFile(path string) (Model, tea.Cmd) {
	m.selectedFile = path
	content, err := os.ReadFile(path)
	if err != nil {
		return m, nil
	}
	m.content = strings.Split(string(content), "\n")

	m.applyFilter()
	m.state = StateViewing
	return m, nil
}

// Applique le filtre de texte et la coloration syntaxique des logs
func (m *Model) applyFilter() {
	var builder strings.Builder
	filter := strings.ToLower(m.textInput.Value())

	for _, line := range m.content {
		lineLower := strings.ToLower(line)
		lineUpper := strings.ToUpper(line)

		if filter != "" && !strings.Contains(lineLower, filter) {
			continue
		}

		// Coloration basée sur les mots-clés standards (ERROR, WARN, INFO)
		coloredLine := line
		if strings.Contains(lineUpper, "ERROR") || strings.Contains(lineUpper, "FAIL") || strings.Contains(lineUpper, "CRIT") || strings.Contains(lineUpper, "FATAL") {
			coloredLine = errorStyle.Render(line)
		} else if strings.Contains(lineUpper, "WARN") {
			coloredLine = warnStyle.Render(line)
		} else if strings.Contains(lineUpper, "INFO") || strings.Contains(lineUpper, "DEBUG") || strings.Contains(lineUpper, "NOTICE") {
			coloredLine = infoStyle.Render(line)
		}

		builder.WriteString(coloredLine + "\n")
	}
	m.viewport.SetContent(builder.String())
	m.viewport.GotoTop()
}