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

// Styles locaux
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Pink
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true) // Rouge
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true) // Orange
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))            // Cyan
	
	// Nouveaux styles pour le File Picker
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#500aff")).Bold(true)
)

type Model struct {
	state        SessionState
	
	// Composants
	filePicker   filepicker.Model
	viewport     viewport.Model
	textInput    textinput.Model // Pour le filtre Log
	pathInput    textinput.Model // Pour taper le chemin manuellement (ctrl+l)
	
	// Données
	selectedFile string
	content      []string 
	width        int
	height       int
	
	// États booléens
	filtering    bool // Vrai si on tape un filtre dans le Viewer
	enteringPath bool // Vrai si on tape un chemin dans le Picker
}

// New initialise le modèle avec la taille actuelle de la fenêtre
func New(w, h int) Model {
	// 1. Config du FilePicker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".log", ".txt", ".go", ".md", ".json", ".yaml", ".conf"}
	fp.CurrentDirectory, _ = os.Getwd()
// --- AJOUT : PERSONNALISATION DES COULEURS ---
    // C'est ici qu'on change les couleurs de la liste des fichiers
    fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))      // Curseur Rose (>)
    fp.Styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Texte sélectionné Rose
    fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))   // Dossiers Cyan
    fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))        // Fichiers Blanc
    // ---------------------------------------------

	fp.Height = h - 8 // On réduit un peu pour laisser la place au header/footer
	fp.ShowHidden = false 

	// 2. Config du TextInput pour le filtre (Viewer)
	tiFilter := textinput.New()
	tiFilter.Placeholder = "Filtrer (ex: /db)..."
	tiFilter.CharLimit = 156
	tiFilter.Width = 30

	// 3. Config du TextInput pour le chemin (Picker - Ctrl+L)
	tiPath := textinput.New()
	tiPath.Placeholder = "/home/user/..."
	tiPath.CharLimit = 256
	tiPath.Width = 50

	// 4. Config du Viewport (Viewer)
	vp := viewport.New(w, h-4)

	return Model{
		state:      StateChooseMethod,
		filePicker: fp,
		textInput:  tiFilter,
		pathInput:  tiPath,
		viewport:   vp,
		width:      w,
		height:     h,
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
		m.filePicker.Height = msg.Height - 8 // Ajustement hauteur picker

	case tea.KeyMsg:
		// Gestion globale de la sortie (si on n'est pas en train d'écrire)
		if !m.filtering && !m.enteringPath && (msg.String() == "ctrl+c") {
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	// Machine à état
	switch m.state {

	// ---------------------------------------------------------
	// 1. CHOIX DE LA MÉTHODE (T/G)
	// ---------------------------------------------------------
	case StateChooseMethod:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q":
				return m, func() tea.Msg { return BackMsg{} }
			case "t": // Terminal
				m.state = StatePickingFile
				m.filePicker.CurrentDirectory, _ = os.Getwd()
				return m, m.filePicker.Init()
			case "g": // Graphique
				path, err := zenity.SelectFile(zenity.Filename(m.filePicker.CurrentDirectory + "/"))
				if err == nil && path != "" {
					return m.loadFile(path)
				}
			}
		}

	// ---------------------------------------------------------
	// 2. EXPLORATEUR DE FICHIERS (FILE PICKER)
	// ---------------------------------------------------------
	case StatePickingFile:
		// CAS A : On est en train de taper un chemin manuellement
		if m.enteringPath {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "enter":
					// On valide le chemin
					newPath := m.pathInput.Value()
					info, err := os.Stat(newPath)
					// Si c'est un dossier valide, on y va
					if err == nil && info.IsDir() {
						m.filePicker.CurrentDirectory = newPath
						m.filePicker.Init() // Rafraîchir la liste
					}
					// Quoi qu'il arrive, on quitte le mode input
					m.enteringPath = false
					m.pathInput.Blur()
				case "esc":
					// Annuler
					m.enteringPath = false
					m.pathInput.Blur()
				}
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

		// CAS B : Navigation normale dans le File Picker
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc","q":
				m.state = StateChooseMethod // Retour choix méthode
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

	// ---------------------------------------------------------
	// 3. VISUALISATION DU LOG (VIEWER)
	// ---------------------------------------------------------
	case StateViewing:
		switch msg := msg.(type) {
		case tea.KeyMsg:
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

			switch msg.String() {
			case "/":
				m.filtering = true
				m.textInput.Focus()
				return m, textinput.Blink
			case "backspace":
				m.textInput.Reset()
				m.applyFilter()
			case "esc", "q":
				// Retour au file picker si on veut changer de fichier, ou quitter
				// Ici on fait "Retour au file picker" pour fluidifier
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
		
		// En-tête : Chemin actuel (PWD)
		currentDir := fmt.Sprintf(" %s", pathStyle.Render(m.filePicker.CurrentDirectory))
		
		// 2. Gestion de l'affichage Input ou Picker
		var content string
		if m.enteringPath {
			content = fmt.Sprintf("\n\nEntrez le chemin absolu :\n%s", m.pathInput.View())
		} else {
			content = "\n" + m.filePicker.View()
		}

		// 3. Pied de page : Aide contextuelle grise
		var helpText string
		if m.enteringPath {
			helpText = "enter: valider • esc: annuler"
		} else {
			// Indicateur visuel pour les fichiers cachés
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

// loadFile lit le fichier et prépare le viewport
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

// applyFilter génère le contenu coloré et filtré
func (m *Model) applyFilter() {
	var builder strings.Builder
	filter := strings.ToLower(m.textInput.Value())

	for _, line := range m.content {
		lineLower := strings.ToLower(line)
		lineUpper := strings.ToUpper(line)

		// 1. Filtrage
		if filter != "" && !strings.Contains(lineLower, filter) {
			continue
		}

		// 2. Coloriage
		coloredLine := line
		
		// ERREURS
		if strings.Contains(lineUpper, "ERROR") || strings.Contains(lineUpper, "FAIL") || strings.Contains(lineUpper, "CRIT") || strings.Contains(lineUpper, "FATAL") {
			coloredLine = errorStyle.Render(line)
		// WARNINGS
		} else if strings.Contains(lineUpper, "WARN") { 
			coloredLine = warnStyle.Render(line)
		// INFOS
		} else if strings.Contains(lineUpper, "INFO") || strings.Contains(lineUpper, "DEBUG") || strings.Contains(lineUpper, "NOTICE") {
			coloredLine = infoStyle.Render(line)
		}

		builder.WriteString(coloredLine + "\n")
	}
	m.viewport.SetContent(builder.String())
	m.viewport.GotoTop()
}
